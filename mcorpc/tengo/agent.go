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

	_, err = createRequestMap(req, s)
	if err != nil {
		abortAction(fmt.Sprintf("Could not set rpc data: %s", err), agent, reply)
		return
	}

	_, err = createReplyMap(req.Action, agent, s)
	if err != nil {
		abortAction(fmt.Sprintf("Could not create reply data: %s", err), agent, reply)
		return
	}

	_, err = createDataMap(req, s)
	if err != nil {
		abortAction(fmt.Sprintf("Could not set request data: %s", err), agent, reply)
		return
	}

	c, err := s.Compile()
	if err != nil {
		abortAction(fmt.Sprintf("Could not compile action: %s", err), agent, reply)
		return
	}

	err = c.RunContext(ctx)
	if err != nil {
		abortAction(fmt.Sprintf("Could not run: %s", err), agent, reply)
		return
	}

	repObj, ok := c.Get("reply").Object().(*Reply)
	if !ok {
		abortAction("Could not retrieve reply data", agent, reply)
	}

	reply.Data = repObj.Data()
}

func createDataMap(req *mcorpc.Request, s *script.Script) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})
	err = json.Unmarshal(req.Data, &data)
	if err != nil {
		return nil, err
	}

	err = s.Add("request", data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func createRequestMap(req *mcorpc.Request, s *script.Script) (request map[string]interface{}, err error) {
	request = make(map[string]interface{})

	j, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, &request)
	if err != nil {
		return nil, err
	}

	err = s.Add("rpc", request)
	if err != nil {
		return nil, err
	}

	return request, nil
}

// creates a map[string]interface{} from the DDL setting defaults etc
func createReplyMap(action string, a *mcorpc.Agent, s *script.Script) (reply *Reply, err error) {
	addl, err := agent.Find(a.Name(), []string{"/etc/choria/tengo"})
	if err != nil {
		return nil, fmt.Errorf("could not find DDL for %s: %s", a.Name(), err)
	}

	actInterface, err := addl.ActionInterface(action)
	if err != nil {
		return nil, fmt.Errorf("failed to load action %s#%s: %s", a.Name(), action, err)
	}

	reply = &Reply{
		data: make(map[string]interface{}),
	}

	for item, opts := range actInterface.Output {
		reply.data[item] = opts.Default
	}

	err = s.Add("reply", reply)
	if err != nil {
		return nil, err
	}

	return reply, err
}

func abortAction(reason string, agent *mcorpc.Agent, reply *mcorpc.Reply) {
	agent.Log.Error(reason)
	reply.Statuscode = mcorpc.Aborted
	reply.Statusmsg = reason
}
