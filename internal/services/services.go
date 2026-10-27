package services

// common maps well-known TCP ports to their usual service names.
var common = map[int]string{
	21:   "ftp",
	22:   "ssh",
	23:   "telnet",
	25:   "smtp",
	53:   "dns",
	80:   "http",
	110:  "pop3",
	135:  "msrpc",
	139:  "netbios-ssn",
	143:  "imap",
	443:  "https",
	445:  "microsoft-ds",
	3306: "mysql",
	3389: "rdp",
	5432: "postgresql",
	6379: "redis",
	8080: "http-proxy",
}

// Name returns the usual service name for port, or "-" if unknown.
func Name(port int) string {
	if name, ok := common[port]; ok {
		return name
	}
	return "-"
}
