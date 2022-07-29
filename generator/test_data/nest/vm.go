package nest

type Vm struct {
	Ip string `json:"ip" param:"path,ip"`
	Port int `json:"port" param:"path,port"`
	NetWorks []NetWork
}

type NetWork struct {
	Mark string
	Ns string
}