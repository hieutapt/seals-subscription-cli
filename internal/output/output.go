package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

// JSON pretty-prints raw JSON bytes.
func JSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// SubscriptionList renders a list of subscriptions as a table.
func SubscriptionList(data []byte) error {
	var resp struct {
		Success bool `json:"success"`
		Payload []struct {
			ID        int     `json:"id"`
			Email     string  `json:"email"`
			FirstName string  `json:"first_name"`
			LastName  string  `json:"last_name"`
			Status    string  `json:"status"`
			Currency  string  `json:"currency"`
			Total     float64 `json:"total_value"`
			Interval  string  `json:"delivery_interval"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return printRaw(data)
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader([]string{"ID", "Email", "Name", "Status", "Interval", "Total", "Currency"}),
	)

	for _, s := range resp.Payload {
		_ = table.Append(
			strconv.Itoa(s.ID),
			s.Email,
			s.FirstName+" "+s.LastName,
			s.Status,
			s.Interval,
			fmt.Sprintf("%.2f", s.Total),
			s.Currency,
		)
	}
	return table.Render()
}

// SubscriptionDetail renders a single subscription.
func SubscriptionDetail(data []byte) error {
	var resp struct {
		Success bool `json:"success"`
		Payload struct {
			ID        int     `json:"id"`
			Email     string  `json:"email"`
			FirstName string  `json:"first_name"`
			LastName  string  `json:"last_name"`
			Status    string  `json:"status"`
			Currency  string  `json:"currency"`
			Total     float64 `json:"total_value"`
			Interval  string  `json:"delivery_interval"`
			Billing   string  `json:"billing_interval"`
			CardBrand string  `json:"card_brand"`
			CardLast  string  `json:"card_last_digits"`
			CardExpM  string  `json:"card_expiry_month"`
			CardExpY  string  `json:"card_expiry_year"`
			Address   string  `json:"s_address1"`
			City      string  `json:"s_city"`
			Country   string  `json:"s_country"`
			Items     []struct {
				ID    int     `json:"id"`
				Title string  `json:"title"`
				Qty   int     `json:"quantity"`
				Price string  `json:"price"`
				Final float64 `json:"final_price"`
			} `json:"items"`
			Attempts []struct {
				ID     int    `json:"id"`
				Date   string `json:"date"`
				Status string `json:"status"`
			} `json:"billing_attempts"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return printRaw(data)
	}
	s := resp.Payload

	fmt.Printf("\n  Subscription #%d\n", s.ID)
	fmt.Printf("  %-20s %s\n", "Status:", s.Status)
	fmt.Printf("  %-20s %s\n", "Customer:", s.FirstName+" "+s.LastName+" <"+s.Email+">")
	fmt.Printf("  %-20s %s\n", "Address:", s.Address+", "+s.City+", "+s.Country)
	fmt.Printf("  %-20s %.2f %s\n", "Total:", s.Total, s.Currency)
	fmt.Printf("  %-20s %s\n", "Delivery Interval:", s.Interval)
	fmt.Printf("  %-20s %s\n", "Billing Interval:", s.Billing)
	if s.CardBrand != "" {
		fmt.Printf("  %-20s %s **** %s (exp %s/%s)\n", "Payment:", s.CardBrand, s.CardLast, s.CardExpM, s.CardExpY)
	}

	if len(s.Items) > 0 {
		fmt.Println("\n  Items:")
		t := tablewriter.NewTable(os.Stdout,
			tablewriter.WithHeader([]string{"Item ID", "Title", "Qty", "Price", "Final Price"}),
		)
		for _, item := range s.Items {
			_ = t.Append(
				strconv.Itoa(item.ID),
				item.Title,
				strconv.Itoa(item.Qty),
				item.Price,
				fmt.Sprintf("%.2f", item.Final),
			)
		}
		_ = t.Render()
	}

	if len(s.Attempts) > 0 {
		fmt.Println("\n  Upcoming Billing Attempts:")
		t := tablewriter.NewTable(os.Stdout,
			tablewriter.WithHeader([]string{"Attempt ID", "Date", "Status"}),
		)
		for _, a := range s.Attempts {
			status := a.Status
			if status == "" {
				status = "scheduled"
			}
			_ = t.Append(strconv.Itoa(a.ID), a.Date, status)
		}
		_ = t.Render()
	}
	fmt.Println()
	return nil
}

// ProfileList renders a table of profiles.
func ProfileList(headers []string, rows [][]string) {
	t := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader(headers),
	)
	for _, row := range rows {
		_ = t.Append(row[0], row[1], row[2])
	}
	_ = t.Render()
}

// Success prints a success message from a generic API response.
func Success(data []byte, fallback string) error {
	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &resp); err != nil || resp.Message == "" {
		fmt.Println("✓", fallback)
		return nil
	}
	fmt.Println("✓", resp.Message)
	return nil
}

func printRaw(data []byte) error {
	_, err := os.Stdout.Write(data)
	return err
}
