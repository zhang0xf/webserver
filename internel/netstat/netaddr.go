package netstat

type NetAddr struct {
	addr string
}

func (netAddr *NetAddr) Network() string { return "tcp" }

func (netAddr *NetAddr) SetAddr(addr string) { netAddr.addr = addr }

func (netAddr *NetAddr) String() string { return netAddr.addr }

func (netAddr *NetAddr) GetAddr() string { return netAddr.addr }
