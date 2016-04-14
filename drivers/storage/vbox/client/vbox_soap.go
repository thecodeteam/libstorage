package client

import "encoding/xml"

type envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName xml.Name `xml:"Body"`
		Payload []byte   `xml:",innerxml"`
	}
}

type logonRequest struct {
	XMLName  xml.Name `xml:"IWebsessionManager_logon"`
	Username string   `xml:"username,omitempty"`
	Password string   `xml:"password,omitempty"`
}

type logonResponse struct {
	XMLName   xml.Name `xml:"IWebsessionManager_logonResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type findMachineRequest struct {
	XMLName  xml.Name `xml:"IVirtualBox_findMachine"`
	VbID     string   `xml:"_this,omitempty"`
	NameOrID string   `xml:"nameOrId,omitempty"`
}

type findMachineResponse struct {
	XMLName   xml.Name `xml:"IVirtualBox_findMachineResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type getMachineIDRequest struct {
	XMLName xml.Name `xml:"IMachine_getId"`
	Mobref  string   `xml:"_this,omitempty"`
}

type getMachineIDResponse struct {
	XMLName   xml.Name `xml:"IMachine_getIdResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type getMachineNameRequest struct {
	XMLName xml.Name `xml:"IMachine_getName"`
	Mobref  string   `xml:"_this,omitempty"`
}

type getMachineNameResponse struct {
	XMLName   xml.Name `xml:"IMachine_getNameResponse"`
	Returnval string   `xml:"returnval,omitempty"`
}

type getMachinesRequest struct {
	XMLName xml.Name `xml:"IVirtualBox_getMachines"`
	VbID    string   `xml:"_this,omitempty"`
}

type getMachinesResponse struct {
	XMLName   xml.Name `xml:"IVirtualBox_getMachinesResponse"`
	Returnval []string `xml:"returnval,omitempty"`
}
