package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hieutapt/seals-subscription-cli/internal/api"
	"github.com/hieutapt/seals-subscription-cli/internal/output"
)

func subscriptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription",
		Aliases: []string{"sub", "s"},
		Short:   "Manage subscriptions",
	}

	cmd.AddCommand(subListCmd())
	cmd.AddCommand(subGetCmd())
	cmd.AddCommand(subCancelCmd())
	cmd.AddCommand(subPauseCmd())
	cmd.AddCommand(subReactivateCmd())
	cmd.AddCommand(subResumeCmd())
	cmd.AddCommand(subEditCmd())

	return cmd
}

// ─── List ────────────────────────────────────────────────────────────────────

func subListCmd() *cobra.Command {
	var (
		query         string
		page          int
		activeOnly    bool
		pausedOnly    bool
		cancelledOnly bool
		withItems     bool
		withBilling   bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			data, err := client.ListSubscriptions(api.ListSubscriptionsParams{
				Query:               query,
				Page:                page,
				ActiveOnly:          activeOnly,
				PausedOnly:          pausedOnly,
				CancelledOnly:       cancelledOnly,
				WithItems:           withItems,
				WithBillingAttempts: withBilling,
			})
			if err != nil {
				return err
			}

			if jsonOut {
				return output.JSON(data)
			}
			return output.SubscriptionList(data)
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "Search by email, first name, or last name")
	cmd.Flags().IntVarP(&page, "page", "p", 1, "Page number")
	cmd.Flags().BoolVar(&activeOnly, "active", false, "Show active subscriptions only")
	cmd.Flags().BoolVar(&pausedOnly, "paused", false, "Show paused subscriptions only")
	cmd.Flags().BoolVar(&cancelledOnly, "cancelled", false, "Show cancelled subscriptions only")
	cmd.Flags().BoolVar(&withItems, "with-items", false, "Include subscription items")
	cmd.Flags().BoolVar(&withBilling, "with-billing", false, "Include billing attempts")

	return cmd
}

// ─── Get ─────────────────────────────────────────────────────────────────────

func subGetCmd() *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 && len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("invalid subscription ID: %s", args[0])
				}
				id = n
			}
			if id == 0 {
				return fmt.Errorf("subscription ID is required (use --id or pass as argument)")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			data, err := client.GetSubscription(id)
			if err != nil {
				return err
			}

			if jsonOut {
				return output.JSON(data)
			}
			return output.SubscriptionDetail(data)
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Cancel ──────────────────────────────────────────────────────────────────

func subCancelCmd() *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:   "cancel [id]",
		Short: "Cancel a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			data, err := client.CancelSubscription(id)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(data)
			}
			return output.Success(data, fmt.Sprintf("Subscription #%d cancelled.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Pause ───────────────────────────────────────────────────────────────────

func subPauseCmd() *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:   "pause [id]",
		Short: "Pause a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			data, err := client.PauseSubscription(id)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(data)
			}
			return output.Success(data, fmt.Sprintf("Subscription #%d paused.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Reactivate ──────────────────────────────────────────────────────────────

func subReactivateCmd() *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:   "reactivate [id]",
		Short: "Reactivate a cancelled subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			data, err := client.ReactivateSubscription(id)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(data)
			}
			return output.Success(data, fmt.Sprintf("Subscription #%d reactivated.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Resume ──────────────────────────────────────────────────────────────────

func subResumeCmd() *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:   "resume [id]",
		Short: "Resume a paused subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient()
			if err != nil {
				return err
			}
			data, err := client.ResumeSubscription(id)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(data)
			}
			return output.Success(data, fmt.Sprintf("Subscription #%d resumed.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Edit ────────────────────────────────────────────────────────────────────

func subEditCmd() *cobra.Command {
	var (
		id               int
		deliveryInterval string
		billingInterval  string
		minCycles        int
		maxCycles        int
		deliveryPrice    float64
		deliveryTitle    string
		address1         string
		address2         string
		city             string
		zip              string
		country          string
		countryCode      string
		province         string
		provinceCode     string
		phone            string
		company          string
		firstName        string
		lastName         string
	)

	cmd := &cobra.Command{
		Use:   "edit [id]",
		Short: "Edit a subscription",
		Long: `Edit subscription fields. Only the flags you provide will be sent.

Examples:
  seal-cli subscription edit 12345 --delivery-interval "2 week"
  seal-cli subscription edit --id 12345 --min-cycles 3 --max-cycles 12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}

			edit := map[string]interface{}{}
			if cmd.Flags().Changed("delivery-interval") {
				edit["delivery_interval"] = deliveryInterval
			}
			if cmd.Flags().Changed("billing-interval") {
				edit["billing_interval"] = billingInterval
			}
			if cmd.Flags().Changed("min-cycles") {
				edit["billing_min_cycles"] = minCycles
			}
			if cmd.Flags().Changed("max-cycles") {
				edit["billing_max_cycles"] = maxCycles
			}
			if cmd.Flags().Changed("delivery-price") {
				edit["delivery_price"] = deliveryPrice
			}
			if cmd.Flags().Changed("delivery-title") {
				edit["delivery_method_title"] = deliveryTitle
			}
			if cmd.Flags().Changed("address1") {
				edit["s_address1"] = address1
			}
			if cmd.Flags().Changed("address2") {
				edit["s_address2"] = address2
			}
			if cmd.Flags().Changed("city") {
				edit["s_city"] = city
			}
			if cmd.Flags().Changed("zip") {
				edit["s_zip"] = zip
			}
			if cmd.Flags().Changed("country") {
				edit["s_country"] = country
			}
			if cmd.Flags().Changed("country-code") {
				edit["s_country_code"] = countryCode
			}
			if cmd.Flags().Changed("province") {
				edit["s_province"] = province
			}
			if cmd.Flags().Changed("province-code") {
				edit["s_province_code"] = provinceCode
			}
			if cmd.Flags().Changed("phone") {
				edit["s_phone"] = phone
			}
			if cmd.Flags().Changed("company") {
				edit["s_company"] = company
			}
			if cmd.Flags().Changed("first-name") {
				edit["s_first_name"] = firstName
			}
			if cmd.Flags().Changed("last-name") {
				edit["s_last_name"] = lastName
			}

			if len(edit) == 0 {
				return fmt.Errorf("no fields specified to edit. Use --help to see available flags")
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			data, err := client.EditSubscription(id, edit)
			if err != nil {
				return err
			}
			if jsonOut {
				return output.JSON(data)
			}
			fields := make([]string, 0, len(edit))
			for k := range edit {
				fields = append(fields, k)
			}
			return output.Success(data, fmt.Sprintf("Subscription #%d updated (%s).", id, strings.Join(fields, ", ")))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	cmd.Flags().StringVar(&deliveryInterval, "delivery-interval", "", "Delivery interval (e.g. \"2 week\")")
	cmd.Flags().StringVar(&billingInterval, "billing-interval", "", "Billing interval (e.g. \"4 week\")")
	cmd.Flags().IntVar(&minCycles, "min-cycles", 0, "Minimum billing cycles")
	cmd.Flags().IntVar(&maxCycles, "max-cycles", 0, "Maximum billing cycles")
	cmd.Flags().Float64Var(&deliveryPrice, "delivery-price", 0, "Delivery price")
	cmd.Flags().StringVar(&deliveryTitle, "delivery-title", "", "Delivery method title")
	cmd.Flags().StringVar(&firstName, "first-name", "", "Shipping first name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Shipping last name")
	cmd.Flags().StringVar(&address1, "address1", "", "Shipping address line 1")
	cmd.Flags().StringVar(&address2, "address2", "", "Shipping address line 2")
	cmd.Flags().StringVar(&city, "city", "", "Shipping city")
	cmd.Flags().StringVar(&zip, "zip", "", "Shipping ZIP")
	cmd.Flags().StringVar(&country, "country", "", "Shipping country")
	cmd.Flags().StringVar(&countryCode, "country-code", "", "Shipping country code (e.g. US)")
	cmd.Flags().StringVar(&province, "province", "", "Shipping province")
	cmd.Flags().StringVar(&provinceCode, "province-code", "", "Shipping province code (e.g. NY)")
	cmd.Flags().StringVar(&phone, "phone", "", "Shipping phone")
	cmd.Flags().StringVar(&company, "company", "", "Shipping company")

	return cmd
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func resolveID(flagVal int, args []string, name string) (int, error) {	if flagVal != 0 {
		return flagVal, nil
	}
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return 0, fmt.Errorf("invalid %s ID: %s", name, args[0])
		}
		return n, nil
	}
	return 0, fmt.Errorf("%s ID is required (use --id or pass as argument)", name)
}


