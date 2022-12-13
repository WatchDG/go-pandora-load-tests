package main

import (
	"github.com/spf13/afero"
	"github.com/valyala/fasthttp"
	PandoraCli "github.com/yandex/pandora/cli"
	PandoraCore "github.com/yandex/pandora/core"
	PandoraCoreAggregatorNetSample "github.com/yandex/pandora/core/aggregator/netsample"
	PandoraCoreImport "github.com/yandex/pandora/core/import"
	PandoraCoreRegister "github.com/yandex/pandora/core/register"
	"log"
)

type HttpAmmo struct {
	Tag     string
	Url     string
	Method  string
	Headers map[string]string
	Body    string
}

func NewHttpAmmo() PandoraCore.Ammo {
	return &HttpAmmo{}
}

type GunConfig struct{}

type HttpGun struct {
	client     fasthttp.Client
	aggregator PandoraCore.Aggregator
	PandoraCore.GunDeps
}

func NewHttpGun(config GunConfig) *HttpGun {
	return &HttpGun{
		client: fasthttp.Client{},
	}
}

func (g *HttpGun) Bind(aggregator PandoraCore.Aggregator, deps PandoraCore.GunDeps) error {
	g.aggregator = aggregator
	g.GunDeps = deps
	return nil
}

func (g *HttpGun) Shoot(ammo PandoraCore.Ammo) {
	customAmmo := ammo.(*HttpAmmo)
	g.shoot(customAmmo)
}

func (g *HttpGun) shoot(ammo *HttpAmmo) {
	request := fasthttp.AcquireRequest()
	request.SetRequestURI(ammo.Url)
	request.Header.SetMethod(ammo.Method)
	for headerName, headerValue := range ammo.Headers {
		request.Header.Set(headerName, headerValue)
	}
	if ammo.Method == "POST" {
		request.SetBody([]byte(ammo.Body))
	}

	response := fasthttp.AcquireResponse()

	sample := PandoraCoreAggregatorNetSample.Acquire(ammo.Tag)

	if err := g.client.Do(request, response); err != nil {
		log.Panic(err)
	}

	sample.SetProtoCode(response.StatusCode())

	defer func() {
		g.aggregator.Report(sample)
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
}

func main() {
	fs := afero.NewOsFs()
	PandoraCoreImport.Import(fs)

	PandoraCoreImport.RegisterCustomJSONProvider("httpAmmo", NewHttpAmmo)
	PandoraCoreRegister.Gun("httpGun", NewHttpGun)

	PandoraCli.Run()
}
