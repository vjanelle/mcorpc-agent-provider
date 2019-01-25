package tengo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/choria-io/go-protocol/protocol"
	"github.com/choria-io/mcorpc-agent-provider/mcorpc"
	"github.com/d5/tengo/script"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	os.Setenv("MCOLLECTIVE_CERTNAME", "rip.mcollective")
	RegisterFailHandler(Fail)
	RunSpecs(t, "McoRPC/Tengo")
}

var _ = Describe("McoRPC/Tengo", func() {
	It("Should do stuff", func() {
		req := &mcorpc.Request{
			Action:     "install",
			Agent:      "package",
			CallerID:   "choria=rip.mcollective",
			Collective: "mcollective",
			Data:       []byte(`{"hello":"world"}`),
			Filter:     protocol.NewFilter(),
			RequestID:  "123",
			SenderID:   "test.example.net",
			TTL:        60,
			Time:       time.Now(),
		}

		j, err := json.Marshal(req)
		Expect(err).ToNot(HaveOccurred())

		reqmap := make(map[string]interface{})
		err = json.Unmarshal(j, &reqmap)
		Expect(err).ToNot(HaveOccurred())

		datmap := make(map[string]interface{})
		err = json.Unmarshal(req.Data, &datmap)
		Expect(err).ToNot(HaveOccurred())

		replymap := make(map[string]interface{})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		code, err := ioutil.ReadFile("testdata/agents/package/install.tengo")
		Expect(err).ToNot(HaveOccurred())

		s := script.New(code)
		err = s.Add("rpc", reqmap)
		Expect(err).ToNot(HaveOccurred())
		err = s.Add("request", datmap)
		Expect(err).ToNot(HaveOccurred())
		err = s.Add("reply", replymap)
		Expect(err).ToNot(HaveOccurred())

		c, err := s.Compile()
		Expect(err).ToNot(HaveOccurred())

		err = c.RunContext(ctx)
		Expect(err).ToNot(HaveOccurred())

		fmt.Printf("%#v\n", c.Get("reply").Map())
	})
})
