package gofc

import (
	"fmt"
	"net"

	"github.com/naoki9911/gofc/ofprotocol/ofp13"
)

var DEFAULT_PORT = 6653

type OFController struct {
	listener *net.TCPListener
	conns    []*net.TCPConn
}

func NewOFController() *OFController {
	controller := &OFController{
		listener: nil,
		conns:    make([]*net.TCPConn, 0),
	}
	return controller
}

func (c *OFController) ServerLoop(listenPort int) {
	var port int

	if listenPort <= 0 || listenPort >= 65536 {
		fmt.Println("Invalid port was specified. listen port must be between 0 - 65535.")
		return
	}
	port = listenPort

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", port))
	c.listener, err = net.ListenTCP("tcp", tcpAddr)

	ofc := NewSimpleOFController()
	GetAppManager().RegistApplication(ofc)

	if err != nil {
		return
	}

	// wait for connect from switch
	for {
		conn, err := c.listener.AcceptTCP()
		if err != nil {
			return
		}
		c.conns = append(c.conns, conn)
		go handleConnection(conn)
	}
}

func (c *OFController) Stop() {
	err := c.listener.Close()
	fmt.Println(err)

	for idx := range c.conns {
		err = c.conns[idx].Close()
		fmt.Println(err)
	}

	c.listener = nil
	c.conns = make([]*net.TCPConn, 0)
}

/**
 *
 */
func handleConnection(conn *net.TCPConn) {
	// send hello
	hello := ofp13.NewOfpHello()
	_, err := conn.Write(hello.Serialize())
	if err != nil {
		fmt.Println(err)
	}

	// create datapath
	dp := NewDatapath(conn)

	// launch goroutine
	go dp.recvLoop()
	go dp.sendLoop()
}

/**
 * basic controller
 */
type SimpleOFController struct {
	echoInterval int32 // echo interval
}

func NewSimpleOFController() *SimpleOFController {
	ofc := new(SimpleOFController)
	ofc.echoInterval = 60
	return ofc
}

// func (c *OFController) HandleHello(msg *ofp13.OfpHello, dp *Datapath) {
// 	fmt.Println("recv Hello")
// 	// send feature request
// 	featureReq := ofp13.NewOfpFeaturesRequest()
// 	Send(dp, featureReq)
// }

func (c *SimpleOFController) HandleSwitchFeatures(msg *ofp13.OfpSwitchFeatures, dp *Datapath) {
	fmt.Println("recv SwitchFeatures")
	// handle FeatureReply
	dp.datapathId = msg.DatapathId
}

func (c *SimpleOFController) HandleEchoRequest(msg *ofp13.OfpHeader, dp *Datapath) {
	// send EchoReply
	echo := ofp13.NewOfpEchoReply()
	(*dp).Send(echo)
}

func (c *SimpleOFController) ConnectionUp() {
	// handle connection up
}

func (c *SimpleOFController) ConnectionDown() {
	// handle connection down
}

func (c *SimpleOFController) sendEchoLoop() {
	// send echo request forever
}
