package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	var targetUrl string
	fmt.Print("Hedef Site URL'sini giriniz (Örn: https://web-scraper.com): ")
	_, err := fmt.Scanln(&targetUrl)
	if err != nil && err.Error() != "unexpected newline" {
		log.Fatalf("Giriş hatası: %v", err)
	}

	if targetUrl == "" {
		fmt.Println("Hata: Bir URL girmelisiniz.")
		return
	}
	parsedUrl, err := url.Parse(targetUrl)
	if err != nil {
		log.Fatalf("Url formatı hatalı: %v", err)
	}
	hostname := parsedUrl.Hostname()
	if hostname == "" {
		hostname = "bilinmeyen_site"
	}

	domainName := strings.ReplaceAll(hostname, ".", "_")

	htmlFileName := domainName + ".html"

	pngFileName := domainName + ".png"

	linksFileName := domainName + "_links.txt"

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1920, 1080),
		chromedp.DisableGPU,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var htmlContent string
	var screenshotBuf []byte
	var links []string

	fmt.Println("\n--- AŞAMA 1: HTML Bilgisi ---")
	fmt.Printf("Scraper başlatılıyor: %s\n", targetUrl)
	fmt.Printf("[*] Bağlantı kuruluyor: %s/\n", targetUrl)
	fmt.Printf("Dosyalar '%s' domain adıyla kaydedilecektir.\n", domainName)

	tasks := chromedp.Tasks{
		chromedp.Navigate(targetUrl),
		chromedp.Sleep(2 * time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			fmt.Println("[+] Bağlantı başarılı! (200 OK)")
			return nil
		}),
		chromedp.OuterHTML("html", &htmlContent),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &links),
	}

	err = chromedp.Run(ctx, tasks)
	if err != nil {
		log.Fatalf("Hata oluştu (Siteye ulaşılamadı veya zaman aşımı): %v", err)
	}

	if err := os.WriteFile(htmlFileName, []byte(htmlContent), 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("[+] Sayfanın ham HTML içeriği '%s' dosyasına kaydedildi.\n", htmlFileName)

	fmt.Println("\n--- AŞAMA 2: Ekran Görüntüsü Alma ---")
	fmt.Println("[*] Chrome başlatılıyor, lütfen bekleyiniz...")
	fmt.Println("-> Siteye gidiliyor ve görüntü işleniyor...")

	err = chromedp.Run(ctx,
		chromedp.FullScreenshot(&screenshotBuf, 90),
	)

	if err != nil {
		fmt.Printf("[-] Ekran görüntüsü alma hatası: %v\n", err)
	} else {
		if err := os.WriteFile(pngFileName, screenshotBuf, 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("[+] Ekran görüntüsü başarıyla '%s' dosyasına kaydedildi.\n", pngFileName)
	}
	linkFile, err := os.Create(linksFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer linkFile.Close()

	for _, link := range links {
		if link != "" {
			linkFile.WriteString(link + "\n")
		}
		
	}
	fmt.Printf("[+] Bulunan %d adet URL '%s' dosyasına kaydedildi.\n", len(links), linksFileName)

		fmt.Println("\n[√] Tüm görevler tamamlandı.")
}
