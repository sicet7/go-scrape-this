package scrape

import (
	"context"
	_ "embed"
	"encoding/base64"
	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
	"time"
)

//go:embed ScrapeVehicle.js
var scrapeVehicleScript string

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func ScrapeVehicle(searchType string, value string, timeout time.Duration) (map[string]interface{}, error) {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, timeout)
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
		chromedp.Evaluate(scrapeVehicleScript, &res),
	)
	if err != nil {
		return map[string]interface{}{}, err
	}

	output = res
	output["vehicle_image"] = toBase64(image)

	res = map[string]interface{}{}
	image = []byte{}

	err = chromedp.Run(ctx,
		chromedp.Click("#li-visKTTabset-1 a"),
		chromedp.WaitReady("#visKTTabset"),
		chromedp.FullScreenshot(&image, 90),
		chromedp.Evaluate(scrapeVehicleScript, &res),
	)

	if err != nil {
		return map[string]interface{}{}, err
	}

	output = lo.Assign(output, res)
	output["technical_details_image"] = toBase64(image)

	return output, nil
}
