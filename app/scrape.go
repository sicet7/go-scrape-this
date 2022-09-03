package app

import (
	"context"
	_ "embed"
	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"log"
)

//go:embed getData.js
var script string

func ScarpeVehicle(searchType string, value string) map[string]interface{} {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var output = map[string]interface{}{}
	var res = map[string]interface{}{}
	var image []byte
	chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")

	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://motorregister.skat.dk/dmr-kerne/koeretoejdetaljer/visKoeretoej`),
		chromedp.WaitReady(searchType),
		chromedp.Click(searchType),
		chromedp.SetValue("#soegeord", value),
		chromedp.Submit("#searchForm"),
		chromedp.WaitReady("#visKTTabset"),
		chromedp.FullScreenshot(&image, 90),
		chromedp.Evaluate(script, &res),
	)
	if err != nil {
		log.Fatal(err)
	}

	output = res
	output["vehicle_image"] = toBase64(image)

	res = map[string]interface{}{}
	image = []byte{}

	err = chromedp.Run(ctx,
		chromedp.Click("#li-visKTTabset-1 a"),
		chromedp.WaitReady("#visKTTabset"),
		chromedp.FullScreenshot(&image, 90),
		chromedp.Evaluate(script, &res),
	)

	if err != nil {
		log.Fatal(err)
	}

	output = lo.Assign(output, res)
	output["technical_details_image"] = toBase64(image)

	return output
}
