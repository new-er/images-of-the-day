package sources

type ChannelResult[T any]  struct{
	Value T
	Err   error
}
