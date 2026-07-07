package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	CMC             float64      `json:"cmc"`
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

type MTGItem struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name"`
	Colors          []string `json:"colors"`
	Type            string   `json:"type"`
	Set             string   `json:"set"`
	SetName         string   `json:"set_name"`
	OracleText      *string  `json:"oracle_text"`
	FlavorText      *string  `json:"flavor_text"`
	Rarity          string   `json:"rarity"`
	CollectorNumber int      `json:"collector_number"`
	Artist          string   `json:"artist"`
	ReleasedAt      string   `json:"released_at"`
	Image           string   `json:"image"`
	Back            *string  `json:"back"`
}

type ScryfallCardSelection map[string]MTGItem

var magicColors = map[string]string{
	"W": "White",
	"U": "Blue",
	"B": "Black",
	"R": "Red",
	"G": "Green",
}

// escapeText replaces newline characters with their escaped equivalent (\n)
// to ensure text can be safely stored in single-line formats or databases.
func escapeText(text string) string {
	return strings.ReplaceAll(text, "\n", "\\n")
}

// escapeText, magicColors, and the Scryfall response types above are shared with
// SearchCard, the live lookup path used by the mtg enricher.

// SearchCard fetches a card directly from Scryfall by set code and collector number.
// DOCS: https://scryfall.com/docs/api/cards/collector
func SearchCard(name string, set string, number int) (ScryfallCardSelection, error) {
	// Direct lookup by set+number — more reliable than search query parsing.
	cardURL := fmt.Sprintf(
		"https://api.scryfall.com/cards/%s/%d",
		strings.TrimSpace(strings.ToLower(set)),
		number,
	)

	req, err := http.NewRequest("GET", cardURL, nil)
	if err != nil {
		return nil, fmt.Errorf("(SearchCard): failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Rivendell/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("(SearchCard): request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		log.Printf("[SearchCard] Error body: %s", string(errBody))
		return nil, fmt.Errorf("(SearchCard): %d - %s | %s/%d",
			resp.StatusCode,
			resp.Status,
			set,
			number,
		)
	}

	var card ScryfallCardData
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, fmt.Errorf("(SearchCard): failed to decode response: %w", err)
	}

	// Handle oracle text
	var oracleText *string
	if card.OracleText != "" {
		escaped := escapeText(card.OracleText)
		oracleText = &escaped
	} else if len(card.CardFaces) > 0 && card.CardFaces[0].OracleText != "" {
		escaped := escapeText(card.CardFaces[0].OracleText)
		oracleText = &escaped
	}

	// Handle flavor text
	var flavorText *string
	if card.FlavorText != "" {
		escaped := escapeText(card.FlavorText)
		flavorText = &escaped
	} else if len(card.CardFaces) > 0 && card.CardFaces[0].FlavorText != "" {
		escaped := escapeText(card.CardFaces[0].FlavorText)
		flavorText = &escaped
	}

	// Handle colors
	var colors []string
	for _, color := range card.Colors {
		if colorName, ok := magicColors[color]; ok {
			colors = append(colors, colorName)
		} else {
			colors = append(colors, color)
		}
	}

	// Handle front image
	var image string
	if card.ImageUris.Png != "" {
		image = card.ImageUris.Png
	} else if len(card.CardFaces) > 0 && card.CardFaces[0].ImageUris.Png != "" {
		image = card.CardFaces[0].ImageUris.Png
	}

	// Handle back image
	var back *string
	if len(card.CardFaces) > 1 && card.CardFaces[1].ImageUris.Png != "" {
		backImg := card.CardFaces[1].ImageUris.Png
		back = &backImg
	}

	collectorNumber, _ := strconv.Atoi(card.CollectorNumber)

	item := MTGItem{
		Name:            card.Name,
		Colors:          colors,
		Type:            card.TypeLine,
		Set:             strings.ToUpper(card.Set),
		SetName:         card.SetName,
		OracleText:      oracleText,
		FlavorText:      flavorText,
		Rarity:          card.Rarity,
		CollectorNumber: collectorNumber,
		Artist:          card.Artist,
		ReleasedAt:      card.ReleasedAt,
		Image:           image,
		Back:            back,
	}

	key := fmt.Sprintf("%s - (%s) #%d", card.Name, strings.ToUpper(card.Set), collectorNumber)
	selection := ScryfallCardSelection{key: item}

	return selection, nil
}
