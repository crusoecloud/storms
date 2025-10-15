package utils

type CmdFactory struct {
	AdminClientProvider  AdminClientProvider
	StorMSClientProvider StorMSClientProvider
}

func NewCmdFactory(
	adminClientProvider AdminClientProvider,
	stormsClientProvider StorMSClientProvider,

) *CmdFactory {
	return &CmdFactory{
		AdminClientProvider:  adminClientProvider,
		StorMSClientProvider: stormsClientProvider,
	}
}
