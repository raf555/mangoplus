// package mangoplus provides a client for using MangaPlus API.
//
// Usage:
//
//	import "github.com/raf555/mangoplus"
//
//	client, err := mangoplus.NewClient()
//	// handle error
//	secret, err := client.Register(ctx)
//
// [Client.Register] call is required unless you specify the secret via [WithSecret] option or using [WithAutoRegister] option.
// Example:
//
//	client, err := mangoplus.NewClient(mangoplus.WithSecret("some_secret"))
//
// or
//
//	client, err := mangoplus.NewClient(mangoplus.WithAutoRegister(context.Background()))
//
// Disclaimer: this package is an unofficial API wrapper for the MangaPlus android
// application and is not affiliated with, endorsed by, or sponsored by Shueisha
// or MangaPlus. "MangaPlus" and all related content are trademarks of their
// respective owners. The API is undocumented and may change or break at any time.
// Use of this package may be subject to MangaPlus's Terms of Service; users are
// responsible for ensuring their use complies with applicable terms and laws.
package mangoplus
