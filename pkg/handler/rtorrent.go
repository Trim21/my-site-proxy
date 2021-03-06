package handler

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mrobinsn/go-rtorrent/xmlrpc"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"app/pkg/logger"
	rt "app/pkg/rtorrent"
)

func setupRTorrentMetrics(router fiber.Router) {
	entryPoint, found := os.LookupEnv("RTORRENT_API_ENTRYPOINT")
	if !found {
		return
	}

	_, err := url.Parse(entryPoint)
	if err != nil {
		logger.Fatal("can't parse RTORRENT_API_ENTRYPOINT", zap.String("value", entryPoint))
	}

	rpc := xmlrpc.NewClient(entryPoint, true)

	router.Get("/rtorrent/metrics", func(ctx *fiber.Ctx) error {
		logger.Debug("export rtorrent metrics")
		v, err := rt.GetGlobalData(rpc)
		if err != nil {
			return errors.WithMessage(err, "rpc")
		}

		fmt.Fprintf(ctx, "rtorrent_upload_total_bytes{hostname=%s} %d\n", strconv.Quote(v.Hostname), v.UpTotal)
		fmt.Fprintf(ctx, "rtorrent_download_total_bytes{hostname=%s} %d\n", strconv.Quote(v.Hostname), v.DownTotal)

		for i := range v.Torrents {
			writeRtorrentTorrent(ctx, &v.Torrents[i])
		}

		return nil
	})
}

func writeRtorrentTorrent(w io.Writer, t *rt.Torrent) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "# torrent", strconv.Quote(t.Name))
	fmt.Fprintln(w, "# label:", t.Label)

	if t.Label == "" {
		fmt.Fprintf(w, "rtorrent_torrent_download_total_bytes{hash=%s} %d\n",
			strconv.Quote(t.Hash), t.DownloadTotal)
	} else {
		for _, label := range t.Labels() {
			fmt.Fprintf(w, "rtorrent_torrent_download_bytes{label=%s, hash=%s} %d\n",
				strconv.Quote(label), strconv.Quote(t.Hash), t.DownloadTotal)

			fmt.Fprintf(w, "rtorrent_torrent_upload_bytes{label=%s, hash=%s} %d\n",
				strconv.Quote(label), strconv.Quote(t.Hash), t.UploadTotal)
		}
	}
}
