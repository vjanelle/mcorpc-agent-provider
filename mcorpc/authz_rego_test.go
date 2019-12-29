package mcorpc

import (
	"encoding/json"

	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/go-choria/server/agents"
	"github.com/choria-io/go-config"
	"github.com/choria-io/go-protocol/filter/facts"
	"github.com/choria-io/go-testutil"
	"github.com/choria-io/mcorpc-agent-provider/mcorpc/golang/rpcutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("regoPolicy", func() {
	var (
		authz  *regoPolicy
		logger *logrus.Entry
		fw     *choria.Framework
		cn     *testutil.ChoriaNetwork
		err    error
		am     *agents.Manager
	)

	BeforeEach(func() {

		logger = logrus.NewEntry(logrus.New())
		logger.Logger.SetLevel(logrus.DebugLevel)
		logger.Logger.Out = GinkgoWriter

		cfg := config.NewConfigForTests()
		cfg.ClassesFile = "testdata/rego/classes.txt"
		cfg.ConfigFile = "testdata/server.conf"
		cfg.FactSourceFile = "testdata/rego/facts.json"

		cfg.DisableSecurityProviderVerify = true

		fw, err = choria.NewWithConfig(cfg)
		Expect(err).ToNot(HaveOccurred())

		cn, err = testutil.StartChoriaNetwork(cfg)
		Expect(err).ToNot(HaveOccurred())

		am = agents.New(nil, fw, nil, cn.ServerInstance(), nil)
		rpc, err := rpcutil.New(am)

		Expect(err).ToNot(HaveOccurred())

		rpc.SetServerInfo(cn.ServerInstance())

		metadata := agents.Metadata{
			Name:    "ginkgo",
			Author:  "stub@example.com",
			License: "Apache-2.0",
			Timeout: 10,
			URL:     "https://choria.io",
			Version: "1.0.0",
		}

		authz = &regoPolicy{
			cfg: cfg,
			log: logger,
			req: &Request{
				Agent:    "ginkgo",
				Action:   "boop",
				CallerID: "choria=ginkgo.mcollective",
			},
			agent: &Agent{
				meta:             &metadata,
				Log:              logger,
				Config:           cfg,
				Choria:           fw,
				ServerInfoSource: cn.ServerInstance(),
			},
		}

	})

	AfterEach(func() {
		cn.Stop()
	})

	Describe("Basic tests", func() {
		var (
		// defaultFacts = json.RawMessage(`{"stub": true, "buts": "big"}`)
		)

		BeforeEach(func() {
			// is.EXPECT().Facts().Return(json.RawMessage(defaultFacts)).AnyTimes()
			// is.EXPECT().Classes().Return([]string{"alpha", "beta"}).AnyTimes()
		})

		Context("When the user agent or caller is right", func() {
			It("Should succeed", func() {
				auth, err := authz.authorize()
				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeTrue())
			})

			It("Default policy should fail", func() {
				authz.agent.meta.Name = "boop"
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeFalse())
			})

		})

		Context("When facts are correct", func() {
			It("Should succeed", func() {

				authz.agent.meta.Name = "facts"
				auth, err := authz.authorize()
				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeTrue())

			})
		})

		Context("When classes are present and available", func() {
			It("Should succeed", func() {
				authz.agent.meta.Name = "classes"
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeTrue())
			})
		})
	})

	Describe("Failing tests", func() {
		var (
		// defaultFacts = json.RawMessage(`{"stub": false, "buts": true}`)
		)
		BeforeEach(func() {
			// is.EXPECT().Facts().Return(json.RawMessage(defaultFacts)).AnyTimes()
			// is.EXPECT().Classes().Return([]string{"charlie", "delta"}).AnyTimes()
		})
		Context("When the user agent or caller is wrong", func() {
			It("Should fail if agent isn't ginkgo", func() {
				authz.req.CallerID = "not=it"
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeFalse())
			})

			It("Should fail with a default policy", func() {
				authz.req.CallerID = "not=it"
				authz.agent.meta.Name = "boop"
				Expect(authz.agent.Name()).To(Equal("boop"))

				authz.cfg.SetOption("plugin.regopolicy.enable_default", "y")
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeFalse())
			})
		})

		Context("When the facts don't line up", func() {
			It("Should fail", func() {
				authz.agent.meta.Name = "facts"
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeFalse())
			})
		})

		Context("When classes are not what we expect", func() {
			It("Should fail", func() {
				authz.agent.meta.Name = "classes"
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeFalse())
			})
		})
	})

	Describe("Agents", func() {
		var (
		// defaultFacts = json.RawMessage(`{"stub": true, "buts": "big"}`)
		)

		BeforeEach(func() {
			// is.EXPECT().Facts().Return(json.RawMessage(defaultFacts)).AnyTimes()
			// is.EXPECT().Classes().Return([]string{"alpha", "beta"}).AnyTimes()
		})

		Context("If agent exists on the server", func() {
			It("Should succed", func() {
				authz.agent.meta.Name = "agent"
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeTrue())
			})
		})
	})

	Describe("Facts", func() {

		BeforeEach(func() {
			// is.EXPECT().Classes().Return([]string{"alpha", "beta"}).AnyTimes()

		})

		Context("If facts are empty it should fail", func() {
			It("Should fail", func() {
				f, err := facts.JSON("", logger)
				Expect(err).To(HaveOccurred())
				Expect(f).To(BeEquivalentTo(json.RawMessage("{}")))

				// is.EXPECT().Facts().Return(f).AnyTimes()

				authz.agent.meta.Name = "Facts"
				auth, err := authz.authorize()

				Expect(err).ToNot(HaveOccurred())
				Expect(auth).To(BeFalse())
			})
		})
	})

	/*
	* I am not exactly sure what it means to be "in provisioning" as far as agents goes
	 */

	/*
		Describe("Provision mode", func() {
			var (
				defaultFacts = json.RawMessage(`{"stub": true, "buts": "big"}`)
			)

			BeforeEach(func() {
				is.EXPECT().Facts().Return(json.RawMessage(defaultFacts)).AnyTimes()
				is.EXPECT().Classes().Return([]string{"alpha", "beta"}).AnyTimes()
			})

			Context("It should only auth if provisioning is set to true", func() {
				It("Should succeed", func() {
					authz.agent.meta.Name = "provisioning"
					auth, err := authz.authorize()

					Expect(err).ToNot(HaveOccurred())
					Expect(auth).To(BeTrue())
				})

			})
		})
	*/
})
