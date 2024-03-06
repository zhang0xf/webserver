package conn

type IOPumper interface {
	writePump(conn *Conn)
	readPump(conn *Conn)
}
