package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Preview struct {
	Source      string `json:"source"`
	PreviewedAt string `json:"previewed_at"`
	SourceURI   string `json:"source_uri"`
}

type Legalities struct {
	Gladiator       string `json:"gladiator"`
	Historicbrawl   string `json:"historicbrawl"`
	Explorer        string `json:"explorer"`
	Vintage         string `json:"vintage"`
	Oldschool       string `json:"oldschool"`
	Legacy          string `json:"legacy"`
	Pauper          string `json:"pauper"`
	Standard        string `json:"standard"`
	Modern          string `json:"modern"`
	Penny           string `json:"penny"`
	Brawl           string `json:"brawl"`
	Duel            string `json:"duel"`
	Paupercommander string `json:"paupercommander"`
	Premodern       string `json:"premodern"`
	Alchemy         string `json:"alchemy"`
	Future          string `json:"future"`
	Commander       string `json:"commander"`
	Historic        string `json:"historic"`
	Pioneer         string `json:"pioneer"`
}

type RelatedUris struct {
	TcgplayerInfiniteDecks    string `json:"tcgplayer_infinite_decks"`
	Edhrec                    string `json:"edhrec"`
	Gatherer                  string `json:"gatherer"`
	TcgplayerInfiniteArticles string `json:"tcgplayer_infinite_articles"`
}

type PurchaseUris struct {
	Cardhoarder string `json:"cardhoarder"`
	Cardmarket  string `json:"cardmarket"`
	Tcgplayer   string `json:"tcgplayer"`
}

type ImageUris struct {
	Large      string `json:"large"`
	BorderCrop string `json:"border_crop"`
	Normal     string `json:"normal"`
	ArtCrop    string `json:"art_crop"`
	Small      string `json:"small"`
	Png        string `json:"png"`
}

type CardFaces struct {
	Object         string    `json:"object"`
	Name           string    `json:"name"`
	ManaCost       string    `json:"mana_cost"`
	TypeLine       string    `json:"type_line"`
	OracleText     string    `json:"oracle_text"`
	FlavorText     string    `json:"flavor_text"`
	Colors         []string  `json:"colors"`
	Power          string    `json:"power"`
	Toughness      string    `json:"toughness"`
	Artist         string    `json:"artist"`
	ArtistID       string    `json:"artist_id"`
	IllustrationID string    `json:"illustration_id"`
	ImageUris      ImageUris `json:"image_uris"`
}

type Prices struct {
	Tix       string `json:"tix"`
	UsdEtched string `json:"usd_etched"`
	Eur       string `json:"eur"`
	EurFoil   string `json:"eur_foil"`
	Usd       string `json:"usd"`
	UsdFoil   string `json:"usd_foil"`
}

type ScryfallCardData struct {
	Rarity          string       `json:"rarity"`
	Artist          string       `json:"artist"`
	Frame           string       `json:"frame"`
	Power           string       `json:"power"`
	URI             string       `json:"uri"`
	ID              string       `json:"id"`
	TcgplayerID     int          `json:"tcgplayer_id"`
	Digital         bool         `json:"digital"`
	CMC             int          `json:"cmc"`
	PennyRank       int          `json:"penny_rank"`
	Preview         Preview      `json:"preview"`
	CollectorNumber string       `json:"collector_number"`
	Layout          string       `json:"layout"`
	SetID           string       `json:"set_id"`
	FullArt         bool         `json:"full_art"`
	Nonfoil         bool         `json:"nonfoil"`
	Textless        bool         `json:"textless"`
	BorderColor     string       `json:"border_color"`
	SetURI          string       `json:"set_uri"`
	Finishes        []string     `json:"finishes"`
	SetSearchURI    string       `json:"set_search_uri"`
	Legalities      Legalities   `json:"legalities"`
	IllustrationID  string       `json:"illustration_id"`
	Games           []string     `json:"games"`
	OracleID        string       `json:"oracle_id"`
	OracleText      string       `json:"oracle_text"`
	ImageStatus     string       `json:"image_status"`
	Reserved        bool         `json:"reserved"`
	MtgoID          int          `json:"mtgo_id"`
	ManaCost        string       `json:"mana_cost"`
	PrintsSearchURI string       `json:"prints_search_uri"`
	Colors          []string     `json:"colors"`
	Name            string       `json:"name"`
	CardmarketID    int          `json:"cardmarket_id"`
	RelatedUris     RelatedUris  `json:"related_uris"`
	CardBackID      string       `json:"card_back_id"`
	Oversized       bool         `json:"oversized"`
	ScryfallSetURI  string       `json:"scryfall_set_uri"`
	ColorIdentity   []string     `json:"color_identity"`
	TypeLine        string       `json:"type_line"`
	PurchaseUris    PurchaseUris `json:"purchase_uris"`
	Object          string       `json:"object"`
	ScryfallURI     string       `json:"scryfall_uri"`
	SetName         string       `json:"set_name"`
	EdhrecRank      int          `json:"edhrec_rank"`
	MultiverseIDs   []int        `json:"multiverse_ids"`
	Set             string       `json:"set"`
	Foil            bool         `json:"foil"`
	ReleasedAt      string       `json:"released_at"`
	RulingsURI      string       `json:"rulings_uri"`
	Toughness       string       `json:"toughness"`
	ImageUris       ImageUris    `json:"image_uris"`
	CardFaces       []CardFaces  `json:"card_faces"`
	Promo           bool         `json:"promo"`
	Booster         bool         `json:"booster"`
	StorySpotlight  bool         `json:"story_spotlight"`
	SetType         string       `json:"set_type"`
	Variation       bool         `json:"variation"`
	Keywords        []string     `json:"keywords"`
	ArtistIDs       []string     `json:"artist_ids"`
	FlavorText      string       `json:"flavor_text"`
	Prices          Prices       `json:"prices"`
	HighresImage    bool         `json:"highres_image"`
	Lang            string       `json:"lang"`
	Reprint         bool         `json:"reprint"`
}

type CleanMTG struct {
	Name            string
	Colors          *[]string
	Type            string
	Set             string
	SetName         string
	OracleText      *string
	FlavorText      *string
	Rarity          string
	CollectorNumber int
	Artist          string
	ReleasedAt      string
	Image           string
	Back            *string
}

var magicColors = map[string]string{
	"W": "White",
	"U": "Blue",
	"B": "Black",
	"R": "Red",
	"G": "Green",
}

func escapeText(text string) string {
	return strings.ReplaceAll(text, "\n", "\\n")
}

func parseMTGURL(url string) (string, error) {
	regex, err := regexp.Compile(`cards/([a-f0-9\-]+)/oembed`)
	if err != nil {
		return "", fmt.Errorf("[parseMTGURL][regexp.Compile]: %w", err)
	}

	matches := regex.FindStringSubmatch(url)
	if len(matches) == 2 {
		return matches[1], nil
	}

	return "", fmt.Errorf("[parseMTGURL]: No matches found%w", nil)
}

func getOembedURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("[getOembedURL]: %w", err)
	}

	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("[getOembedURL]: %w", err)
	}

	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "head" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "link" {
					isAlternate := false
					isOembed := false
					var href string

					for _, a := range c.Attr {
						if a.Key == "rel" && a.Val == "alternate" {
							isAlternate = true
						} else if a.Key == "type" && a.Val == "application/json+oembed" {
							isOembed = true
						} else if a.Key == "href" {
							href = a.Val
						}
					}

					if isAlternate && isOembed && href != "" {
						return href
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			result := f(c)
			if result != "" {
				return result
			}
		}

		return ""
	}

	return f(doc), nil
}

func mapCardData(data ScryfallCardData) (CleanMTG, error) {
	var oText, fText string

	if data.OracleText != "" {
		oText = escapeText(data.OracleText)
	} else if len(data.CardFaces) > 0 && data.CardFaces[0].OracleText != "" {
		oText = escapeText(data.CardFaces[0].OracleText)
	}

	if data.FlavorText != "" {
		fText = escapeText(data.FlavorText)
	} else if len(data.CardFaces) > 0 && data.CardFaces[0].FlavorText != "" {
		fText = escapeText(data.CardFaces[0].FlavorText)
	}

	var colorNames []string
	if len(data.Colors) > 0 {
		for _, color := range data.Colors {
			if name, ok := magicColors[color]; ok {
				colorNames = append(colorNames, name)
			}
		}
	}

	var item CleanMTG
	item.Name = data.Name
	item.Colors = &colorNames
	item.Type = data.TypeLine
	item.Set = strings.ToUpper(data.Set)
	item.SetName = data.SetName
	item.OracleText = &oText
	item.FlavorText = &fText
	item.Rarity = data.Rarity
	collectorNumber, err := strconv.Atoi(data.CollectorNumber)
	if err != nil {
		return CleanMTG{}, fmt.Errorf("[mapCardData][strconv.Atoi]: %w", err)
	}
	item.CollectorNumber = collectorNumber
	item.Artist = data.Artist
	item.ReleasedAt = data.ReleasedAt

	if len(data.CardFaces) > 0 {
		item.Image = data.CardFaces[0].ImageUris.Png
	} else {
		item.Image = data.ImageUris.Png
	}

	if len(data.CardFaces) > 1 {
		item.Back = &data.CardFaces[1].ImageUris.Png
	}

	return item, nil
}

func GetMTGInfo(url string) (CleanMTG, error) {
	link, linkErr := getOembedURL(url)
	if linkErr != nil {
		return CleanMTG{}, fmt.Errorf("[GetMTGInfo]%w", linkErr)
	}

	id, idErr := parseMTGURL(link)
	if idErr != nil {
		return CleanMTG{}, fmt.Errorf("[GetMTGInfo]%w", idErr)
	}

	resp, err := http.Get(fmt.Sprintf("https://api.scryfall.com/cards/%s", id))
	if err != nil {
		return CleanMTG{}, fmt.Errorf("[GetMTGInfo][http.Get]: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = fmt.Errorf("[fetch]: %d - %s (%s)", resp.StatusCode, resp.Status, id)
		return CleanMTG{}, fmt.Errorf("[GetMTGInfo]%w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CleanMTG{}, fmt.Errorf("[GetMTGInfo][io.ReadAll]: %w", err)
	}

	var response ScryfallCardData
	if err := json.Unmarshal(body, &response); err != nil {
		return CleanMTG{}, fmt.Errorf("[GetMTGInfo][json.Unmarshal]: %w", err)
	}

	data, err := mapCardData(response)
	if err != nil {
		return CleanMTG{}, fmt.Errorf("[GetMTGInfo]%w", err)
	}

	return data, nil
}
