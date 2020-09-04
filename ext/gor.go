package ext

import (
    "bufio"
    "bytes"
    "io/ioutil"
    "strconv"
    "strings"

    "net/http"

    "github.com/xsyr/goawk/interp"
)

type gorModule struct {
    rt *interp.Runtime
    funcs []*FuncDef

    payloadSeparator string
    payloadSeparatorAsBytes []byte

    pendingHeader []byte
    raw bytes.Buffer
    typ int64
    id string
    ts int64

    req *http.Request
    resp *http.Response

    body []byte
}

func newGorModule() Module {
    var m = &gorModule{
        payloadSeparator:        "\n🐵🙈🙉\n",
        payloadSeparatorAsBytes: []byte("\n🐵🙈🙉\n"),
    }
    m.define()
    return m
}

func (m *gorModule) Desc() string {
    return `
goreplay - e.g.
    [awk for process gor DISK FILE]:
        $ goreplay --input-raw :80  --output-file v1_image_reqs.gor --http-allow-url '^/container/v1/image' --output-file-append

        function handleGorPayload() {
            method = __gor_this_method();
            if(method == "GET") {
                print __gor_this_path(), __gor_this_param("data");
            } else if(method == "POST") {
                print __gor_this_path(), __gor_this_body();
            }
            #print __gor_this_header("X-Forwarded-For");
        
            __gor_reset();
        }
        
        {
            if(__gor_try_parse(__rawline()) > 0) {
                handleGorPayload()
            }
        }
        END {
            if(__gor_parse() > 0) {
                handleGorPayload()
            }
        }

    [awk for process gor stdout]
    $ goreplay  --input-raw :80 -input-raw-track-response -prettify-http --output-stdout  --http-allow-url '^/container/v1/image'
    
    function handleGor_Req_Payload() {
        method = __gor_this_method();
        if(method == "GET") {
            print "REQ", __gor_this_param("data");
        } else if(method == "POST") {
            print "REQ", __gor_this_body();
        }
        
        headers[] = __gor_this_headers();
        print headers["X-Forwarded-For"];
    
        __gor_reset();
    }
    
    function handleGor_Resp_Payload() {
        print "RESP", __gor_this_body();
        __gor_reset();
    }
    
    {
        if(index($0, "version") > 0){
            # skip goreplay version info
            next;
        }
    
        if(__gor_try_parse(__rawline()) > 0) {
            if(__gor_this_type() == "REQ") {
                handleGor_Req_Payload();
            } else {
                handleGor_Resp_Payload();
            }
        }
    }
    END {
        if(__gor_parse() > 0) {
            if(__gor_this_type() == "REQ") {
                handleGor_Req_Payload();
            } else {
                handleGor_Resp_Payload();
            }
        }
    }


`
}

func (m *gorModule)SetRuntime(rt *interp.Runtime) {
    m.rt = rt
}

func (m *gorModule)payloadMeta(payload []byte) ([][]byte, int) {
    headerSize := bytes.IndexByte(payload, '\n')
    if headerSize < 0 {
        headerSize = 0
    }
    return bytes.Split(payload[:headerSize], []byte{' '}), headerSize
}

func (m *gorModule)isProcessingNextFile(line []byte) bool {
    meta, headerSize := m.payloadMeta(line)
    if headerSize == 0 { return false }
    if len(meta) < 3 { return false }

    t := string(meta[0])
    if t != "1" && t != "2" && t != "3" { return false }
    if len(meta[1]) != 40 { return false }
    return m.raw.Len() > 0
}

func (m *gorModule)parsePayload() int {
    asBytes := m.raw.Bytes()
    if len(asBytes) == 0 { return 0 }
    meta, headerSize := m.payloadMeta(asBytes)

    m.typ, _ = strconv.ParseInt(string(meta[0]), 10, 64)
    m.id = string(meta[1])
    m.ts, _  = strconv.ParseInt(string(meta[2]), 10, 64)

    r := bytes.NewReader(asBytes[headerSize+1:])

    var err error
    if m.typ == 1 {
        m.req, err = http.ReadRequest(bufio.NewReader(r))
        if err != nil {
            return 0
        }
        if m.req.Method == "POST" {
            m.body, _ = ioutil.ReadAll(m.req.Body)
        }
    } else if m.typ == 2 || m.typ == 3 {
        m.resp, err = http.ReadResponse(bufio.NewReader(r), nil)
        if err != nil {
            return 0
        }
        m.body, _ = ioutil.ReadAll(m.resp.Body)
    }
    return 1
}

func (m *gorModule) Funcs() []*FuncDef {
    return m.funcs
}

func (m *gorModule) define() {
    if m.funcs != nil {
        return
    }
    m.funcs = []*FuncDef{
        &FuncDef{
            Desc : "__gor_is_sep(line) int. return if is seperator.",
            Func : func(line []byte) int {
                if bytes.Equal(m.payloadSeparatorAsBytes[1:], line) {
                    return 1;
                }
                return 0
            },
        },
        &FuncDef{
            Desc : "__gor_this_raw() []byte. return raw gor data of this part.",
            Func : func() []byte {
                return m.raw.Bytes()
            },
        },
        &FuncDef{
            Desc : "__gor_this_method() string. GET/POST.",
            Func : func() string {
                if m.req != nil {
                    return m.req.Method
                }
                return ""
            },
        },
        &FuncDef{
            Desc : "__gor_this_type() string. REQ=request, RESP=response, RR=replay response",
            Func : func() string {
                switch m.typ {
                case 1: return "REQ"
                case 2: return "RESP"
                case 3: return "RR"
                }
                return ""
            },
        },
        &FuncDef{
            Desc : "__gor_this_body() string. body data for POST.",
            Func : func() []byte {
                return  m.body
            },
        },
        &FuncDef{
            Desc : "__gor_this_id() string. gor id.",
            Func : func() string {
                return  m.id
            },
        },
        &FuncDef{
            Desc : "__gor_this_url() string.",
            Func : func() string {
                if m.req != nil {
                    return m.req.URL.String()
                }
                return ""
            },
        },
        &FuncDef{
            Desc : "__gor_this_path() string.",
            Func : func() string {
                if m.req != nil {
                    return m.req.URL.Path
                }
                return ""
            },
        },
        &FuncDef{
            Desc : "__gor_this_param(name) string.",
            Func : func(name string) string {
                if m.req != nil {
                    return m.req.URL.Query().Get(name)
                }
                return ""
            },
        },
        &FuncDef{
            Desc : "__gor_this_params() map[string]string.",
            Func : func() map[string]string {
                if m.req == nil {
                    return nil
                }
                ps := make(map[string]string)
                for k, v := range m.req.URL.Query() {
                    if len(v) > 0 {
                        ps[k] = v[0]
                    } else {
                        ps[k] = ""
                    }
                }
                return ps
            },
        },
        &FuncDef{
            Desc : "__gor_this_header(name) string.",
            Func : func(name string) string {
                if m.req != nil {
                    return m.req.Header.Get(name)
                }
                return ""
            },
        },
        &FuncDef{
            Desc : "__gor_this_headers() map[string]string.",
            Func : func() map[string]string {
                if m.req == nil {
                    return nil
                }
                ps := make(map[string]string)
                for k, v := range m.req.Header {
                    if len(v) > 0 {
                        ps[k] = v[0]
                    } else {
                        ps[k] = ""
                    }
                }
                return ps
            },
        },
        &FuncDef{
            Desc : "__gor_reset(). reset the parse result, called after __gor_try_parse() return true.",
            Func : func() string {
                m.raw.Reset()
                m.req = nil
                m.resp = nil
                m.ts = 0
                m.typ = 0
                m.body = nil
                return ""
            },
        },
        &FuncDef{
            Desc : "__gor_parse() int. parse the remain data. 1=parsed success",
            Func : func() int {
                return m.parsePayload()
            },
        },
        &FuncDef{
            Desc : "__gor_try_parse(line) int. if line is seperator or EOF, try parse. 1=parsed success",
            Func : func(line []byte) int {
                if bytes.Equal(m.payloadSeparatorAsBytes[1:], line) {
                    if m.parsePayload() == 0 {
                        return 0
                    }

                    m.raw.Write(line)
                    return 1
                }
                if m.isProcessingNextFile(line) {
                    m.pendingHeader = line
                    return m.parsePayload()
                }
                if len(m.pendingHeader) > 0 {
                    m.raw.Write(m.pendingHeader)
                    m.pendingHeader = nil
                }
                m.raw.Write(line)
                return 0
            },
        },
    }

    for _, fn := range m.funcs {
        fn.Name = strings.TrimSpace(strings.Split(fn.Desc, "(")[0])
    }

}