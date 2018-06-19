package client

import (
	"errors"
	"time"

	"github.com/choria-io/go-choria/choria"
	"github.com/choria-io/go-protocol/protocol"
)

// RequestOptions are options for a RPC request
type RequestOptions struct {
	Targets         []string
	BatchSize       int
	BatchSleep      time.Duration
	ProtocolVersion string
	RequestType     string
	Workers         int
	Collective      string
	ReplyTo         string
	ProcessReplies  bool
	ReceiveReplies  bool
	Replies         chan *choria.ConnectorMessage
	Progress        bool
	Timeout         time.Duration
	Handler         Handler
	RequestID       string

	totalStats *Stats

	// per batch
	stats *Stats

	fw *choria.Framework
}

// RequestOption is a function capable of setting an option
type RequestOption func(*RequestOptions)

// NewRequestOptions creates a initialized request options
func NewRequestOptions(fw *choria.Framework) *RequestOptions {
	return &RequestOptions{
		ProtocolVersion: protocol.RequestV1,
		RequestType:     "direct_request",
		Collective:      fw.Config.MainCollective,
		ProcessReplies:  true,
		ReceiveReplies:  true,
		Progress:        false,
		Timeout:         time.Duration(fw.Config.DiscoveryTimeout+20) * time.Second,
		Workers:         3,
		stats:           NewStats(),
		totalStats:      NewStats(),
		fw:              fw,
	}
}

// ConfigureMessage configures a pre-made message object based on the settings contained
func (o *RequestOptions) ConfigureMessage(msg *choria.Message) error {
	o.totalStats.RequestID = msg.RequestID
	o.RequestID = msg.RequestID

	if len(o.Targets) > 0 {
		msg.DiscoveredHosts = o.Targets
	} else {
		o.Targets = msg.DiscoveredHosts
	}

	o.totalStats.SetDiscoveredNodes(o.Targets)

	msg.SetProtocolVersion(o.ProtocolVersion)

	if o.RequestType == "request" && o.BatchSize > 0 {
		return errors.New("batched mode required direct_request mode")
	}

	err := msg.SetType(o.RequestType)
	if err != nil {
		return err
	}

	if o.BatchSize == 0 {
		o.BatchSize = len(o.Targets)
	}

	stdtarget := choria.ReplyTarget(msg, msg.RequestID)
	if o.ReplyTo == "" {
		o.ReplyTo = stdtarget
	}

	// the reply target is such that we'd probably not receive replies
	// so disable processing replies
	if stdtarget != o.ReplyTo {
		o.ReceiveReplies = false
	}

	err = msg.SetReplyTo(o.ReplyTo)
	if err != nil {
		return err
	}

	err = msg.SetCollective(o.Collective)
	if err != nil {
		return err
	}

	return nil
}

// Stats retrieves the stats for the completed request
func (o *RequestOptions) Stats() *Stats {
	return o.totalStats
}

// WithProgress enable a progress writer
func WithProgress() RequestOption {
	return func(o *RequestOptions) {
		o.Progress = true
	}
}

// Targets configures targets for a RPC request
func Targets(t []string) RequestOption {
	return func(o *RequestOptions) {
		o.Targets = t
	}
}

// Protocol sets the protocol version to use
func Protocol(v string) RequestOption {
	return func(o *RequestOptions) {
		o.ProtocolVersion = v
	}
}

// DirectRequest force the request to be a direct request
func DirectRequest() RequestOption {
	return func(o *RequestOptions) {
		o.RequestType = "direct_request"
	}
}

// BroadcastRequest for the request to be a broadcast mode
//
// **NOTE:** You need to ensure you have filters etc done
func BroadcastRequest() RequestOption {
	return func(o *RequestOptions) {
		o.RequestType = "request"
	}
}

// Workers configures the amount of workers used to process responses
// this is ignored during batched mode as that is always done with a
// single worker
func Workers(w int) RequestOption {
	return func(o *RequestOptions) {
		o.Workers = w
	}
}

// Collective sets the collective to target a message at
func Collective(c string) RequestOption {
	return func(o *RequestOptions) {
		o.Collective = c
	}
}

// ReplyTo sets a custom reply to, else the connector will determine it
func ReplyTo(r string) RequestOption {
	return func(o *RequestOptions) {
		o.ReplyTo = r
		o.ReceiveReplies = false
	}
}

// InBatches performs requests in batches
func InBatches(size int, sleep int) RequestOption {
	return func(o *RequestOptions) {
		o.BatchSize = size
		o.BatchSleep = time.Second * time.Duration(sleep)
		o.Workers = 1
	}
}

// Replies creates a custom channel for replies and will avoid processing them
func Replies(r chan *choria.ConnectorMessage) RequestOption {
	return func(o *RequestOptions) {
		o.Replies = r
		o.ProcessReplies = false
	}
}

// Timeout configures the request timeout
func Timeout(t time.Duration) RequestOption {
	return func(o *RequestOptions) {
		o.Timeout = t
	}
}

// ReplyHandler configures a callback to be called for each message received
func ReplyHandler(f Handler) RequestOption {
	return func(o *RequestOptions) {
		o.Handler = f
	}
}