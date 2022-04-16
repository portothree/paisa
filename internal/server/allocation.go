package server

import (
	"strings"
	"time"

	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"

	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Aggregate struct {
	Date         time.Time `json:"date"`
	Account      string    `json:"account"`
	Amount       float64   `json:"amount"`
	MarketAmount float64   `json:"market_amount"`
}

func GetAllocation(db *gorm.DB) gin.H {
	var postings []posting.Posting
	result := db.Where("account like ?", "Asset:%").Order("date ASC").Find(&postings)
	if result.Error != nil {
		log.Fatal(result.Error)
	}

	now := time.Now()
	postings = lo.Map(postings, func(p posting.Posting, _ int) posting.Posting {
		p.MarketAmount = service.GetMarketPrice(db, p, now)
		return p
	})
	aggregates := computeAggregate(postings, now)
	aggregates_timeline := computeAggregateTimeline(postings)
	return gin.H{"aggregates": aggregates, "aggregates_timeline": aggregates_timeline}
}

func computeAggregateTimeline(postings []posting.Posting) []map[string]Aggregate {
	var timeline []map[string]Aggregate

	var p posting.Posting
	var pastPostings []posting.Posting

	end := time.Now()
	for start := postings[0].Date; start.Before(end); start = start.AddDate(0, 0, 1) {
		for len(postings) > 0 && (postings[0].Date.Before(start) || postings[0].Date.Equal(start)) {
			p, postings = postings[0], postings[1:]
			pastPostings = append(pastPostings, p)
		}

		timeline = append(timeline, computeAggregate(pastPostings, start))
	}
	return timeline
}

func computeAggregate(postings []posting.Posting, date time.Time) map[string]Aggregate {
	byAccount := lo.GroupBy(postings, func(p posting.Posting) string { return p.Account })
	result := make(map[string]Aggregate)
	for account, ps := range byAccount {
		var parts []string
		for _, part := range strings.Split(account, ":") {
			parts = append(parts, part)
			parent := strings.Join(parts, ":")
			result[parent] = Aggregate{Account: parent}
		}

		amount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.Amount }, 0.0)
		marketAmount := lo.Reduce(ps, func(acc float64, p posting.Posting, _ int) float64 { return acc + p.MarketAmount }, 0.0)
		result[account] = Aggregate{Date: date, Account: account, Amount: amount, MarketAmount: marketAmount}

	}
	return result
}
