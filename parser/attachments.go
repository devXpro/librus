package parser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

// DownloadAttachments finds all download buttons on the message page, extracts URLs and downloads the files
// Returns the path to the directory where files were downloaded or empty string if no attachments were found
func DownloadAttachments(ctx context.Context, messageURL string) (string, error) {
	fmt.Println("Looking for download buttons...")

	// Navigate to message page if URL is provided
	if messageURL != "" {
		actions := []chromedp.Action{
			chromedp.Navigate(messageURL),
			logAction("Navigated to message URL"),
			chromedp.Sleep(3 * time.Second),

			// Wait for page to load completely
			chromedp.WaitVisible(`#formWiadomosci`, chromedp.ByQuery),
			logAction("Message page loaded successfully"),
		}

		if err := chromedp.Run(ctx, actions...); err != nil {
			return "", fmt.Errorf("failed to navigate to message page: %w", err)
		}
	}

	// Get browser cookies for using in HTTP requests later
	var cookies []*http.Cookie
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// Get cookies from the browser's network domain
		cookiesData, err := network.GetCookies().Do(ctx)
		if err != nil {
			return err
		}

		for _, cookie := range cookiesData {
			cookies = append(cookies, &http.Cookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Domain:   cookie.Domain,
				Expires:  time.Unix(int64(cookie.Expires), 0),
				Secure:   cookie.Secure,
				HttpOnly: cookie.HTTPOnly,
			})
		}
		fmt.Printf("Extracted %d cookies from browser\n", len(cookies))
		return nil
	}))

	if err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}

	// Get the count of download buttons
	var attachmentCount int
	err = chromedp.Run(ctx, chromedp.Evaluate(`
		document.querySelectorAll('img[src="/assets/img/homework_files_icons/download.png"]').length
	`, &attachmentCount))

	if err != nil {
		return "", fmt.Errorf("failed to count attachments: %w", err)
	}

	if attachmentCount == 0 {
		fmt.Println("No attachments found on this page")
		return "", nil
	}

	fmt.Printf("Found %d attachment(s)\n", attachmentCount)

	// Create a temporary directory with a UUID
	tempDir := filepath.Join("/tmp", uuid.New().String())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	fmt.Printf("Created temporary directory: %s\n", tempDir)

	// For each attachment button, extract URL and download the file
	for i := 0; i < attachmentCount; i++ {
		fmt.Printf("Processing attachment %d of %d\n", i+1, attachmentCount)

		// Get the onclick attribute value for this attachment button
		var onclickValue string
		err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`
			document.querySelectorAll('img[src="/assets/img/homework_files_icons/download.png"]')[%d].getAttribute('onclick')
		`, i), &onclickValue))

		if err != nil {
			fmt.Printf("Failed to get onclick attribute for attachment %d: %v\n", i+1, err)
			continue
		}

		// Extract attachment URL from onclick attribute
		relativeURL, err := extractURLFromOnclick(onclickValue)
		if err != nil {
			fmt.Printf("Failed to extract URL from onclick for attachment %d: %v\n", i+1, err)
			continue
		}

		// Construct full URL
		fullURL := "https://synergia.librus.pl" + relativeURL
		fmt.Printf("Extracted URL: %s\n", fullURL)

		// Get redirect URL using HTTP client instead of browser navigation
		redirectURL, err := getRedirectURL(fullURL, cookies)
		if err != nil {
			fmt.Printf("Failed to get redirect URL for attachment %d: %v\n", i+1, err)
			continue
		}

		// Add /get to the URL to get the direct download link
		downloadURL := redirectURL + "/get"
		fmt.Printf("Download URL: %s\n", downloadURL)

		// Download the file
		filename, err := downloadFile(downloadURL, i+1, cookies, tempDir)
		if err != nil {
			fmt.Printf("Failed to download file for attachment %d: %v\n", i+1, err)
			continue
		}

		fmt.Printf("Successfully downloaded attachment %d to %s\n", i+1, filename)
	}

	return tempDir, nil
}

// getRedirectURL makes a HTTP request to the given URL and returns the redirect URL
// without actually following the redirect
func getRedirectURL(url string, cookies []*http.Cookie) (string, error) {
	// Create a client that doesn't follow redirects
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Return an error to prevent following redirects
			return http.ErrUseLastResponse
		},
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	// Add cookies to the request
	for _, cookie := range cookies {
		// Only add cookies that are relevant for the domain we're requesting
		if strings.Contains(url, cookie.Domain) {
			req.AddCookie(cookie)
		} else {
			// Try to check by hostname
			parsedURL, err := neturl.Parse(url)
			if err == nil && strings.Contains(cookie.Domain, parsedURL.Hostname()) {
				req.AddCookie(cookie)
			}
		}
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Check if we got a redirect status code
	if resp.StatusCode != http.StatusFound &&
		resp.StatusCode != http.StatusMovedPermanently &&
		resp.StatusCode != http.StatusTemporaryRedirect &&
		resp.StatusCode != http.StatusPermanentRedirect {
		return "", fmt.Errorf("expected redirect, got status code: %d", resp.StatusCode)
	}

	// Get the Location header
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no Location header in response")
	}

	// Check if it's a relative URL
	if !strings.HasPrefix(location, "http") {
		// Convert to absolute URL
		baseURL, err := neturl.Parse(url)
		if err != nil {
			return "", err
		}

		relURL, err := neturl.Parse(location)
		if err != nil {
			return "", err
		}

		location = baseURL.ResolveReference(relURL).String()
	}

	return location, nil
}

// extractURLFromOnclick extracts the URL from the onclick attribute
func extractURLFromOnclick(onclick string) (string, error) {
	// Looking for otworz_w_nowym_oknie("URL", "o2", 420, 250)
	// where URL is what we need to extract
	startStr := "otworz_w_nowym_oknie("
	if !strings.Contains(onclick, startStr) {
		return "", fmt.Errorf("onclick does not contain expected function call")
	}

	// Split by the function name
	parts := strings.Split(onclick, startStr)
	if len(parts) < 2 {
		return "", fmt.Errorf("unexpected onclick format")
	}

	// Get the part after the function name
	paramsPart := strings.TrimSpace(parts[1])

	// Try different matching patterns

	// Pattern 1: Double quotes without HTML entities
	if match := strings.Index(paramsPart, `"`); match >= 0 {
		endMatch := strings.Index(paramsPart[match+1:], `"`)
		if endMatch >= 0 {
			url := paramsPart[match+1 : match+1+endMatch]
			url = strings.ReplaceAll(url, `\/`, `/`)
			return url, nil
		}
	}

	// Pattern 2: HTML entity quotes &quot;
	if match := strings.Index(paramsPart, `&quot;`); match >= 0 {
		endMatch := strings.Index(paramsPart[match+6:], `&quot;`)
		if endMatch >= 0 {
			url := paramsPart[match+6 : match+6+endMatch]
			url = strings.ReplaceAll(url, `\/`, `/`)
			return url, nil
		}
	}

	// Pattern 3: Just try to find anything between quotes
	re := regexp.MustCompile(`["']([^"']+)["']`)
	matches := re.FindStringSubmatch(paramsPart)
	if len(matches) > 1 {
		url := matches[1]
		url = strings.ReplaceAll(url, `\/`, `/`)
		return url, nil
	}

	// Last resort: Just look for the URL pattern
	reFallback := regexp.MustCompile(`/wiadomosci/pobierz_zalacznik/\d+/\d+`)
	matches = reFallback.FindStringSubmatch(paramsPart)
	if len(matches) > 0 {
		return matches[0], nil
	}

	// If we're here, the format wasn't as expected
	return "", fmt.Errorf("could not extract URL from onclick attribute")
}

// downloadFile downloads a file from URL and saves it to disk
func downloadFile(url string, index int, cookies []*http.Cookie, targetDir string) (string, error) {
	// Create HTTP client with cookie handling
	client := &http.Client{}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers to mimic browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	// Add cookies to the request
	for _, cookie := range cookies {
		// Only add cookies that are relevant for the domain we're requesting
		if strings.Contains(url, cookie.Domain) {
			req.AddCookie(cookie)
		} else {
			// Try to check by hostname
			parsedURL, err := neturl.Parse(url)
			if err == nil && strings.Contains(cookie.Domain, parsedURL.Hostname()) {
				req.AddCookie(cookie)
			}
		}
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Try to get filename from Content-Disposition header
	filename := ""
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		if strings.Contains(contentDisposition, "filename=") {
			parts := strings.Split(contentDisposition, "filename=")
			if len(parts) > 1 {
				filename = strings.Trim(parts[1], `"'`)
			}
		}
	}

	// If no filename in header, generate one
	if filename == "" {
		// Try to extract from URL
		urlParts := strings.Split(url, "/")
		if len(urlParts) > 0 {
			lastPart := urlParts[len(urlParts)-2] // Use the token part of the URL
			filename = fmt.Sprintf("attachment_%d_%s", index, lastPart)
		} else {
			// Fallback name
			filename = fmt.Sprintf("attachment_%d", index)
		}
	}

	// Clean the filename
	filename = filepath.Base(filename)

	// Full path to save file
	filePath := filepath.Join(targetDir, filename)

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// Write the response body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error writing file: %w", err)
	}

	return filePath, nil
}
