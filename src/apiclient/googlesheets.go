package apiclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsAPI struct {
	ApiKey        string
	SpreadsheetId string
}

var monthname = map[int]string{
	1:  "Январь",
	2:  "Февраль",
	3:  "Март",
	4:  "Апрель",
	5:  "Май",
	6:  "Июнь",
	7:  "Июль",
	8:  "Август",
	9:  "Сентябрь",
	10: "Октябрь",
	11: "Ноябрь",
	12: "Декабрь",
}

type ShopSales struct {
	Name       string
	SalesType  string
	SalesValue float64
}

const ShopsCnt = 5

func (s *SheetsAPI) CallGoogleSheetsApi(day, month int) ([]ShopSales, float64, error) {
	sales := make([]ShopSales, 0, ShopsCnt)
	var month_total float64
	keyBytes, err := base64.StdEncoding.DecodeString(s.ApiKey)
	if err != nil {
		slog.Error("could not base64 decode apikey")
		return nil, 0, err
	}
	config, err := google.JWTConfigFromJSON(keyBytes, sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		slog.Error("unable to create JWT config", "err", err)
		return nil, 0, err
	}

	// Create a Sheets API client using the JWT config.
	ctx := context.Background()
	client := config.Client(ctx)
	sheetsService, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		slog.Error("unable to create Sheets API client", "err", err)
		return nil, 0, err
	}

	for i := 0; i < ShopsCnt; i++ {
		// Define the spreadsheet ID and range of cells to retrieve.
		cellRange := fmt.Sprintf(
			"%s!R%dC%d:R%dC%d",
			monthname[month],
			i*7+3,
			1,
			i*7+3,
			day+2,
		)

		// Make the API request to retrieve the values in the specified cells.
		resp, err := sheetsService.Spreadsheets.Values.Get(s.SpreadsheetId, cellRange).ValueRenderOption("UNFORMATTED_VALUE").Do()
		if err != nil {
			slog.Error("unable to retrieve values", "err", err)
			continue
		}

		if len(resp.Values) == 0 {
			continue
		}

		name := fmt.Sprint(resp.Values[0][0])
		salestype := fmt.Sprint(resp.Values[0][1])
		salesvalue, err := strconv.ParseFloat(fmt.Sprint(resp.Values[0][len(resp.Values[0])-1]), 64)
		if err != nil {
			slog.Error("could not parse salesvalue", "err", err)
			continue
		}
		for i := 2; i < len(resp.Values[0]); i++ {
			salesvalue, err := strconv.ParseFloat(fmt.Sprint(resp.Values[0][i]), 64)
			if err != nil {
				slog.Error("could not parse salesvalue", "err", err)
				month_total = 0
				break
			}
			month_total += salesvalue
		}
		sales = append(sales, ShopSales{name, salestype, salesvalue})
	}
	if len(sales) == 0 {
		return nil, 0, fmt.Errorf("no data")
	}
	return sales, month_total, nil
}
