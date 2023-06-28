package ftpserver

func handleSTRUCommand(conn *FTPServer, args []string) {
	if len(args) == 1 && args[0] == "F" {
		_, _ = conn.Write([]byte("200 Obsolete verb accepted"))
	} else {
		_, _ = conn.Write([]byte("504 Obsolete verb rejected"))
	}
}

func handleMODECommand(conn *FTPServer, args []string) {
	if len(args) == 1 && args[0] == "S" {
		_, _ = conn.Write([]byte("200 Obsolete verb accepted"))
	} else {
		_, _ = conn.Write([]byte("504 Obsolete verb rejected"))
	}
}
