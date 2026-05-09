package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hieutapt/seals-subscription-cli/internal/api"
)

func subscriptionCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription",
		Aliases: []string{"sub", "s"},
		Short:   "Manage subscriptions",
	}

	cmd.AddCommand(subListCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subGetCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subCancelCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subPauseCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subReactivateCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subResumeCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subEditCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subAddItemCmd(profile, jsonOut, agentOut))
	cmd.AddCommand(subRemoveItemCmd(profile, jsonOut, agentOut))

	return cmd
}

// ─── List ────────────────────────────────────────────────────────────────────

func subListCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
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
		Use:     "list",
		Aliases: []string{"ls", "l"},
		Short:   "List subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient(*profile)
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

			return renderSubscriptionList(data, *jsonOut, *agentOut)
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

func subGetCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"show", "g"},
		Short:   "Get details of a subscription",
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

			client, err := newClient(*profile)
			if err != nil {
				return err
			}

			data, err := client.GetSubscription(id)
			if err != nil {
				return err
			}

			return renderSubscriptionDetail(data, *jsonOut, *agentOut)
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Cancel ──────────────────────────────────────────────────────────────────

func subCancelCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:     "cancel [id]",
		Aliases: []string{"cx"},
		Short:   "Cancel a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.CancelSubscription(id)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "cancel", id, fmt.Sprintf("Subscription #%d cancelled.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Pause ───────────────────────────────────────────────────────────────────

func subPauseCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:     "pause [id]",
		Aliases: []string{"p"},
		Short:   "Pause a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.PauseSubscription(id)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "pause", id, fmt.Sprintf("Subscription #%d paused.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Reactivate ──────────────────────────────────────────────────────────────

func subReactivateCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:     "reactivate [id]",
		Aliases: []string{"ra"},
		Short:   "Reactivate a cancelled subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.ReactivateSubscription(id)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "reactivate", id, fmt.Sprintf("Subscription #%d reactivated.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Resume ──────────────────────────────────────────────────────────────────

func subResumeCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var id int

	cmd := &cobra.Command{
		Use:     "resume [id]",
		Aliases: []string{"r"},
		Short:   "Resume a paused subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.ResumeSubscription(id)
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "resume", id, fmt.Sprintf("Subscription #%d resumed.", id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	return cmd
}

// ─── Edit ────────────────────────────────────────────────────────────────────

func subEditCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
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
		Use:     "edit [id]",
		Aliases: []string{"e", "update"},
		Short:   "Edit a subscription",
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

			client, err := newClient(*profile)
			if err != nil {
				return err
			}

			data, err := client.EditSubscription(id, edit)
			if err != nil {
				return err
			}
			fields := make([]string, 0, len(edit))
			for k := range edit {
				fields = append(fields, k)
			}
			return renderSuccess(data, *jsonOut, *agentOut, "edit", id, fmt.Sprintf("Subscription #%d updated (%s).", id, strings.Join(fields, ", ")))
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

// ─── Add Item ────────────────────────────────────────────────────────────────

func subAddItemCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var (
		id               int
		productID        string
		variantID        string
		title            string
		sku              string
		price            float64
		quantity         int
		taxable          bool
		requiresShipping bool
		oneTime          bool
	)

	cmd := &cobra.Command{
		Use:     "add-item [subscription-id]",
		Aliases: []string{"ai"},
		Short:   "Add an item to a subscription",
		Long: `Add a single item to a subscription.

Examples:
  seal-cli subscription add-item 12345 \
    --product-id 4648340258949 --variant-id 32694645424261 \
    --title "Bag of coffee 1kg" --price 24.00 --quantity 1

  seal-cli subscription add-item 12345 \
    --product-id 4648340258949 --variant-id 32694645424261 \
    --title "One-time add-on" --price 9.99 --quantity 1 --one-time`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			if productID == "" {
				return fmt.Errorf("--product-id is required")
			}
			if variantID == "" {
				return fmt.Errorf("--variant-id is required")
			}
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			if price <= 0 {
				return fmt.Errorf("--price is required and must be greater than 0")
			}
			if quantity <= 0 {
				quantity = 1
			}

			item := api.SubscriptionItem{
				ProductID: productID,
				VariantID: variantID,
				Title:     title,
				SKU:       sku,
				Price:     price,
				Quantity:  quantity,
			}
			if taxable {
				item.Taxable = 1
			}
			if requiresShipping {
				item.RequiresShipping = 1
			}
			if oneTime {
				item.OneTime = 1
			}

			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.AddItems(id, []api.SubscriptionItem{item})
			if err != nil {
				return err
			}
			return renderSuccess(data, *jsonOut, *agentOut, "add-item", id, fmt.Sprintf("Item %q added to subscription #%d.", title, id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	cmd.Flags().StringVar(&productID, "product-id", "", "Shopify product ID (required)")
	cmd.Flags().StringVar(&variantID, "variant-id", "", "Shopify variant ID (required)")
	cmd.Flags().StringVar(&title, "title", "", "Product title (required)")
	cmd.Flags().StringVar(&sku, "sku", "", "Product SKU")
	cmd.Flags().Float64Var(&price, "price", 0, "Item price in subscription currency (required)")
	cmd.Flags().IntVar(&quantity, "quantity", 1, "Quantity")
	cmd.Flags().BoolVar(&taxable, "taxable", false, "Item is taxable")
	cmd.Flags().BoolVar(&requiresShipping, "requires-shipping", true, "Item requires shipping")
	cmd.Flags().BoolVar(&oneTime, "one-time", false, "Remove after next renewal (one-time add-on)")

	return cmd
}

// ─── Remove Item ─────────────────────────────────────────────────────────────

func subRemoveItemCmd(profile *string, jsonOut *bool, agentOut *bool) *cobra.Command {
	var (
		id      int
		itemIDs []int
	)

	cmd := &cobra.Command{
		Use:     "remove-item [subscription-id]",
		Aliases: []string{"ri", "rm"},
		Short:   "Remove one or more items from a subscription",
		Long: `Remove items from a subscription by their item IDs.
Use 'seal-cli subscription get <id>' to find item IDs.

Examples:
  seal-cli subscription remove-item 12345 --item-id 1014923
  seal-cli subscription remove-item 12345 --item-id 1014923 --item-id 1014921`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := resolveID(id, args, "subscription")
			if err != nil {
				return err
			}
			if len(itemIDs) == 0 {
				return fmt.Errorf("at least one --item-id is required")
			}

			client, err := newClient(*profile)
			if err != nil {
				return err
			}
			data, err := client.RemoveItems(id, itemIDs)
			if err != nil {
				return err
			}
			ids := make([]string, len(itemIDs))
			for i, id := range itemIDs {
				ids[i] = strconv.Itoa(id)
			}
			return renderSuccess(data, *jsonOut, *agentOut, "remove-item", id, fmt.Sprintf("Item(s) %s removed from subscription #%d.", strings.Join(ids, ", "), id))
		},
	}

	cmd.Flags().IntVar(&id, "id", 0, "Subscription ID")
	cmd.Flags().IntSliceVar(&itemIDs, "item-id", nil, "Item ID to remove (repeat for multiple)")
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


