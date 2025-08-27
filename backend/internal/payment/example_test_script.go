//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/skr1ms/mosaic/internal/payment"
)

// Example script for testing payment API integration
// Run with: go run example_test_script.go
//
// This script demonstrates how to test the payment API endpoints
// Make sure your server is running on the specified URL

func main() {
	// Configuration
	baseURL := "https://photo.doyoupaint.com"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	fmt.Printf("Testing payment integration against: %s\n", baseURL)
	fmt.Println("=" + repeatString("=", 50))

	// Initialize test scenarios
	scenarios := payment.NewTestScenarios(baseURL)

	// Test 1: Integration Test
	fmt.Println("\n🔧 Testing Alfa Bank Integration...")
	integrationResult, err := scenarios.TestIntegrationScenario()
	if err != nil {
		log.Printf("❌ Integration test failed: %v", err)
	} else {
		printTestResult("Integration Test", integrationResult)
	}

	// Test 2: End-to-End Payment Flow
	fmt.Println("\n💳 Testing End-to-End Payment Flow...")
	customerEmail := "test.customer@example.com"
	paymentResult, err := scenarios.EndToEndSuccessfulPayment(customerEmail)
	if err != nil {
		log.Printf("❌ End-to-end test failed: %v", err)
	} else {
		printTestResult("End-to-End Payment", paymentResult)
	}

	// Test 3: Individual API Endpoints
	fmt.Println("\n🛠️ Testing Individual Endpoints...")
	testIndividualEndpoints(baseURL)

	fmt.Println("\n" + repeatString("=", 60))
	fmt.Println("🏁 Testing completed!")
}

func testIndividualEndpoints(baseURL string) {
	helpers := payment.NewTestHelpers(baseURL)

	// Test available options
	fmt.Println("\n📋 Testing /api/payment/options")
	options, err := helpers.GetAvailableOptions()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else {
		fmt.Printf("✅ Success: Found %d sizes, %d styles\n", len(options.Sizes), len(options.Styles))
		printJSON("Options", options)
	}

	// Test integration endpoint
	fmt.Println("\n🔗 Testing /api/payment/test-integration")
	integration, err := helpers.TestIntegration()
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else if integration.Success {
		fmt.Printf("✅ Success: %s\n", integration.Message)
		fmt.Printf("   Config: %s\n", integration.ConfigStatus)
		fmt.Printf("   Test Status: %s\n", integration.TestStatus)
	} else {
		fmt.Printf("⚠️ Integration issues: %s\n", integration.Message)
		if integration.ErrorDetails != "" {
			fmt.Printf("   Details: %s\n", integration.ErrorDetails)
		}
	}

	// Test purchase (this will create a real order, so be careful)
	fmt.Println("\n🛒 Testing /api/payment/purchase")
	purchaseReq := helpers.CreateTestPurchaseRequest(
		"api.test@example.com",
		"40x50",
		"max_colors",
	)

	purchase, err := helpers.TestPurchaseCoupon(purchaseReq)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
	} else if purchase.Success {
		fmt.Printf("✅ Success: Order %s created\n", purchase.OrderNumber)
		fmt.Printf("   Payment URL: %s\n", purchase.PaymentURL)

		// Test order status
		fmt.Printf("\n📊 Testing order status for %s\n", purchase.OrderNumber)
		status, err := helpers.TestGetOrderStatus(purchase.OrderNumber)
		if err != nil {
			fmt.Printf("❌ Failed to get status: %v\n", err)
		} else {
			fmt.Printf("✅ Order status: %s\n", status.Status)
			if status.CouponCode != nil {
				fmt.Printf("   Coupon: %s\n", *status.CouponCode)
			}
		}
	} else {
		fmt.Printf("❌ Purchase failed: %s\n", purchase.Message)
	}
}

func printTestResult(testName string, result *payment.TestScenarioResult) {
	if result.Success {
		fmt.Printf("✅ %s: PASSED\n", testName)
	} else {
		fmt.Printf("❌ %s: FAILED\n", testName)
	}

	fmt.Printf("   Summary: %s\n", result.GetSummary())

	if result.OrderNumber != "" {
		fmt.Printf("   Order: %s\n", result.OrderNumber)
	}

	if result.CouponCode != "" {
		fmt.Printf("   Coupon: %s\n", result.CouponCode)
	}

	// Print step details
	for _, step := range result.Steps {
		status := "❌"
		if step.Success {
			status = "✅"
		}
		fmt.Printf("   %s %s: %s\n", status, step.Description, step.Message)
	}
}

func printJSON(title string, data interface{}) {
	fmt.Printf("\n📋 %s:\n", title)
	jsonData, _ := json.MarshalIndent(data, "   ", "  ")
	fmt.Printf("   %s\n", string(jsonData))
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Usage examples:
//
// Test against local development server:
//   go run example_test_script.go http://localhost:8080
//
// Test against staging:
//   go run example_test_script.go https://staging.doyoupaint.com
//
// Test against production (be careful!):
//   go run example_test_script.go https://photo.doyoupaint.com
//
// The script will:
// 1. Test the Alfa Bank integration endpoint
// 2. Run a complete payment flow simulation
// 3. Test individual API endpoints
// 4. Provide detailed results and suggestions
