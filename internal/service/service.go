package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"OurBear/internal/service/config"

	"github.com/rs/zerolog/log"
)

type chatData struct {
	index int
	order []int
}

type Service struct {
	apiKey     string
	delay      time.Duration
	httpClient http.Client

	mutex     sync.Mutex
	chatsData map[int64]chatData
}

func New(config config.Config) *Service {
	return &Service{
		apiKey: config.ApiKey,
		delay:  config.Delay,
		httpClient: http.Client{
			Timeout: config.Timeout,
		},
		mutex:     sync.Mutex{},
		chatsData: make(map[int64]chatData),
	}
}

func (s *Service) Run(ctx context.Context) error {
	var (
		lastUpdateID int64
		err          error
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if lastUpdateID, err = s.runIteration(ctx, lastUpdateID); err != nil {
			log.Error().
				Err(err).
				Int64("lastUpdateID", lastUpdateID).
				Msg("failed to run iteration")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(s.delay):
		}
	}
}

func (s *Service) runIteration(ctx context.Context, lastUpdateID int64) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", createGetUpdatesURL(lastUpdateID, s.apiKey), nil)
	if err != nil {
		return lastUpdateID, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return lastUpdateID, fmt.Errorf("error making get request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return lastUpdateID, fmt.Errorf("not ok status code: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	var ups updates
	if err = json.NewDecoder(resp.Body).Decode(&ups); err != nil {
		return lastUpdateID, fmt.Errorf("error decoding updates response: %w", err)
	}

	if !ups.Ok {
		return lastUpdateID, fmt.Errorf("didn't get ok in resp body")
	}

	if len(ups.Result) != 0 {
		lastUpdateID = ups.Result[len(ups.Result)-1].UpdateID
	}

	for _, up := range ups.Result {
		if !isGoida(strings.ToLower(up.Message.Text)) {
			continue
		}
		if err = s.sendGoida(ctx, up.Message); err != nil {
			log.Error().
				Err(err).
				Msg("failed to send goida")
		}
	}

	return lastUpdateID, nil
}

func (s *Service) sendGoida(ctx context.Context, message message) error {
	var goidaLink string

	s.mutex.Lock()
	cd, ok := s.chatsData[message.Chat.ID]
	if !ok || cd.index == len(cd.order) {
		cd = chatData{
			index: 0,
			order: makeOrder(),
		}
	}

	goidaLink = goidaLinks[cd.order[cd.index]]
	cd.index++

	s.chatsData[message.Chat.ID] = cd
	s.mutex.Unlock()

	req, err := http.NewRequestWithContext(ctx, "GET", createGoidaURL(s.apiKey, message.Chat.ID, message.MessageID, goidaLink), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making get request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("not ok status code: %v", resp.StatusCode)
	}

	return nil
}

var goidaLinks = []string{
	"https://media1.tenor.com/m/YRRC1UzgBKwAAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D0%B3%D0%BE%D0%BE%D0%BE%D0%BB.gif",
	"https://media1.tenor.com/m/SQAdjmBacesAAAAd/%D0%BC%D0%B5%D0%B4%D0%B2%D0%B5%D0%B4%D1%8C-%D0%B3%D0%BE%D0%BB-%D0%B3%D0%BE%D0%BE%D0%BE%D0%BE%D0%BB.gif",
	"https://media1.tenor.com/m/tDrxZgpCk9cAAAAd/%D0%B4%D0%B0%D0%B9%D1%82%D0%B5-%D0%B3%D0%BE%D0%BB-%D0%B4%D0%B0%D0%B9%D1%82%D0%B5-%D0%B3%D0%BE%D0%BE%D0%BE%D0%BE%D0%BB.gif",
	"https://media1.tenor.com/m/8oouL_By9bAAAAAd/%D0%B3%D0%BE%D0%BE%D0%BE%D0%BB-svo.gif",
	"https://media1.tenor.com/m/k0vJhl9G4NMAAAAd/%D0%BC%D0%B5%D0%B4%D0%B2%D0%B5%D0%B4%D1%8C-z-%D0%B3%D0%BE%D0%BE%D0%BB.gif",
	"https://media1.tenor.com/m/IXe5Lfcr_hkAAAAd/bear-breakcore.gif",
	"https://media1.tenor.com/m/xHEXH8TLgvsAAAAd/raybear.gif",
	"https://media1.tenor.com/m/0uEA7ieXaKoAAAAC/raybear.gif",
	"https://media1.tenor.com/m/YRRC1UzgBKwAAAAC/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D0%B3%D0%BE%D0%BE%D0%BE%D0%BB.gif",
	"https://media1.tenor.com/m/w9VXRl2T-dgAAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D0%B1%D0%BE%D0%B9%D1%81%D1%8F-%D0%BC%D1%8B-%D0%B8%D0%B4%D1%8C%D0%BE%D0%BC.gif",
	"https://media1.tenor.com/m/AH8ePKM3Zm4AAAAC/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/KSnlrs7zBqUAAAAC/%D0%B4%D0%B5%D1%80%D0%B6%D0%B8-%D0%B3%D0%BE%D0%B9%D0%B4%D1%83-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/2zEirW9gj9UAAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D0%B1%D1%80%D0%B0%D1%82%D1%8F-%D1%82%D0%BE%D0%BC%D0%B0%D1%81-%D1%88%D0%B5%D0%BB%D0%B1%D0%B8.gif",
	"https://media1.tenor.com/m/WHhUnkbahX4AAAAC/%D1%83%D1%82%D1%80%D0%BE-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/DnHgYCZIqssAAAAd/goyda-%D0%BE%D1%85%D0%BB%D0%BE%D0%B1%D1%8B%D1%81%D1%82%D0%B8%D0%BD.gif",
	"https://media1.tenor.com/m/fRVLL9GQty8AAAAd/goyda-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/O8PSKseHiCgAAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D1%81%D0%B1%D0%BE%D1%80-%D0%B3%D0%BE%D0%B9%D0%B4%D1%8B.gif",
	"https://media1.tenor.com/m/KvV5NIIqjiMAAAAd/%D0%BF%D1%80%D0%BE%D0%B4%D0%B0%D0%B5%D0%BC-%D0%B3%D0%BE%D0%B9%D0%B4%D1%83-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/ikGQJ03gJq4AAAAd/%D0%BC%D1%8B%D0%B8%D0%B4%D0%B5%D0%BC-%D0%BA%D0%BE%D1%82%D0%B8%D0%BA.gif",
	"https://media1.tenor.com/m/UC8TDXxucKgAAAAd/goida-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/knoJ3ko955gAAAAd/jarvis-goyda.gif",
	"https://media1.tenor.com/m/ASC_CBFskeEAAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D0%BE%D1%85%D0%BB%D0%BE%D0%B1%D1%8B%D1%81%D1%82%D0%B8%D0%BD.gif",
	"https://media1.tenor.com/m/q4WgsT3v8Y4AAAAd/goida.gif",
	"https://media1.tenor.com/m/mVMwrejhwI4AAAAd/%D1%81%D0%B2%D0%BE-%D1%81%D0%BE%D0%B1%D0%B0%D0%BA%D0%B0.gif",
	"https://media1.tenor.com/m/macn07UJ1w0AAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D0%BC%D0%B8%D0%BA%D1%83.gif",
	"https://media1.tenor.com/m/Aj6WQ7arxaUAAAAd/yakuza-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/956gGP7xus8AAAAd/%D0%BF%D0%BE%D1%85%D1%83%D0%B9-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/DNpJ0wbNwk8AAAAd/%D0%BA%D0%B8%D0%B5%D0%B2-%D0%BA%D1%80%D0%B5%D0%BC%D0%BB%D0%B5%D0%B1%D0%BE%D1%82.gif",
	"https://media1.tenor.com/m/zSC3TxBBuC0AAAAd/dog-goal.gif",
	"https://media1.tenor.com/m/RxQOZx_ejxgAAAAd/goooal-goal.gif",
	"https://media1.tenor.com/m/o5l-YLWBDk4AAAAd/snow-team-snow-squad.gif",
	"https://media1.tenor.com/m/r-gzzf2x364AAAAd/touchdown-td.gif",
	"https://media1.tenor.com/m/FuLoJ2M0qjMAAAAd/red-panda-red-pandas.gif",
	"https://media1.tenor.com/m/4CVwlr3LYuAAAAAd/greys-anatomy-grizzly-bear.gif",
	"https://media.tenor.com/pfJe2-vfoPcAAAAj/bear.gif",
	"https://media1.tenor.com/m/6P80PyIyXscAAAAC/blackbeardiner-wink.gif",
	"https://media.tenor.com/EiIgGelxfO0AAAAj/ositos-blancos-osito.gif",
	"https://media1.tenor.com/m/E-T2ZtsOzHgAAAAd/party.gif",
	"https://media1.tenor.com/m/fvILudVo5h4AAAAd/bear-dance.gif",
	"https://media1.tenor.com/m/ak8NMaf6GkkAAAAd/dance-polarbear.gif",
	"https://media.tenor.com/gbBdJjDdkU4AAAAj/petpet-%27pet-pet.gif",
	"https://media.tenor.com/C8-c3HUWARkAAAAj/polar-bear-police-bear.gif",
	"https://media1.tenor.com/m/ND_8Z8BDk-wAAAAd/%D0%BE%D0%B1%D1%8A%D1%8F%D0%B2%D0%BB%D0%B5%D0%BD%D0%B0-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/F2RKuT1DU5YAAAAd/z-goida.gif",
	"https://media1.tenor.com/m/wnKuWB3bzcsAAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-%D1%82%D1%80%D0%B0%D0%BC%D0%BF.gif",
	"https://media1.tenor.com/m/m2Ajo6mQlgkAAAAd/%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0-pootis-engage.gif",
	"https://media1.tenor.com/m/kwSeNk40bF0AAAAd/outis-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
	"https://media1.tenor.com/m/5ZZ9BaEzU7UAAAAd/%D0%B4%D0%B0%D0%B9%D1%82%D0%B5-%D0%B3%D0%BE%D0%B9%D0%B4%D1%83-%D0%B3%D0%BE%D0%B9%D0%B4%D0%B0.gif",
}

func makeOrder() []int {
	order := make([]int, len(goidaLinks))

	for i := range goidaLinks {
		order[i] = i
	}
	rand.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})

	return order
}

type replyParameters struct {
	MessageID int64 `json:"message_id"`
}

func createGoidaURL(apiKey string, chatID, messageID int64, gifURL string) string {
	url := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   fmt.Sprintf("/bot%s/sendAnimation", apiKey),
	}

	rp := replyParameters{MessageID: messageID}
	rpStr, _ := json.Marshal(rp)

	query := url.Query()
	query.Add("chat_id", fmt.Sprint(chatID))
	query.Add("animation", gifURL)
	query.Add("reply_parameters", string(rpStr))
	url.RawQuery = query.Encode()

	return url.String()
}

var golRegexp = regexp.MustCompile(".*го+л.*")

var goidaContainsWords = []string{
	"гойд",
	"медвед",
	"медвеж",
	"русск",
	"россия",
	"рф",
	"слон",
}

func isGoida(s string) bool {
	for _, w := range goidaContainsWords {
		if strings.Contains(s, w) {
			return true
		}
	}
	if golRegexp.MatchString(s) {
		return true
	}

	return false
}

func createGetUpdatesURL(lastUpdateID int64, apiKey string) string {
	url := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   fmt.Sprintf("/bot%s/getUpdates", apiKey),
	}

	query := url.Query()
	query.Set("offset", strconv.FormatInt(lastUpdateID+1, 10))
	url.RawQuery = query.Encode()

	return url.String()
}
