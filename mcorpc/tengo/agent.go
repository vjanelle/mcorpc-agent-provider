package tengo

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/go-choria/server"
	"github.com/choria-io/mcorpc-agent-provider/mcorpc"
	"github.com/choria-io/mcorpc-agent-provider/mcorpc/ddl/agent"
	"github.com/d5/tengo/script"
)

func NewTengoAgent(ddl *agent.DDL, mgr server.AgentManager) (*mcorpc.Agent, error) {
	agent := mcorpc.New(ddl.Metadata.Name, ddl.Metadata, mgr.Choria(), mgr.Logger())

	actions := ddl.ActionNames()
	agent.Log.Debugf("Registering proxy Tengo agent %s with actions: %s", ddl.Metadata.Name, strings.Join(actions, ", "))

	for _, action := range actions {
		int, err := ddl.ActionInterface(action)
		if err != nil {
			return nil, err
		}

		agent.MustRegisterAction(int.Name, tengoAction)
	}

	return agent, nil
}

func tengoAction(ctx context.Context, req *mcorpc.Request, reply *mcorpc.Reply, agent *mcorpc.Agent, conn choria.ConnectorInfo) {
	code, err := ioutil.ReadFile(fmt.Sprintf("/etc/choria/tengo/%s/%s.tengo", req.Agent, req.Action))
	if err != nil {
		abortAction(fmt.Sprintf("Could not read action: %s", err), agent, reply)
		return
	}

	s := script.New(code)
	c, err := s.Compile()
	if err != nil {
		abortAction(fmt.Sprintf("Could not compile action: %s", err), agent, reply)
		return
	}

	err = s.Add("rpc", req)
	if err != nil {
		abortAction(fmt.Sprintf("Could not set rpc data: %s", err), agent, reply)
		return
	}

	replymap := make(map[string]interface{})

	err = s.Add("reply", replymap)
	if err != nil {
		abortAction(fmt.Sprintf("Could not set reply data: %s", err), agent, reply)
		return
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(req.Data, &data)
	if err != nil {
		abortAction(fmt.Sprintf("Could not parse request data: %s", err), agent, reply)
		return
	}

	err = s.Add("request", data)
	if err != nil {
		abortAction(fmt.Sprintf("Could not set request data: %s", err), agent, reply)
		return
	}

	err = c.RunContext(ctx)
	if err != nil {
		abortAction(fmt.Sprintf("Could not run: %s", err), agent, reply)
		return
	}

	reply.Data = c.Get("reply").Map()
}

func abortAction(reason string, agent *mcorpc.Agent, reply *mcorpc.Reply) {
	agent.Log.Error(reason)
	reply.Statuscode = mcorpc.Aborted
	reply.Statusmsg = reason
}
