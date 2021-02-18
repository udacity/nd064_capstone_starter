// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	// "strconv"
	"strings"
	// "time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/genproto"
	// "github.com/GoogleCloudPlatform/microservices-demo/src/frontend/money"
)

type platformDetails struct {
	css      string
	provider string
}

var (
	templates = template.Must(template.New("").
			Funcs(template.FuncMap{
			"renderMoney": renderMoney,
		}).ParseGlob("templates/*.html"))
	plat platformDetails
)

func (fe *frontendServer) homeHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	log.WithField("currency", currentCurrency(r)).Info("home")
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}
	products, err := fe.getProducts(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve products"), http.StatusInternalServerError)
		return
	}
	// cart, err := fe.getCart(r.Context(), sessionID(r))
	// if err != nil {
	// 	renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
	// 	return
	// }

	type productView struct {
		Item  *pb.Product
		Price *pb.Money
	}
	ps := make([]productView, len(products))
	for i, p := range products {
		price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
		if err != nil {
			renderHTTPError(log, r, w, errors.Wrapf(err, "failed to do currency conversion for product %s", p.GetId()), http.StatusInternalServerError)
			return
		}
		ps[i] = productView{p, price}
	}

	//get env and render correct platform banner.
	var env = os.Getenv("ENV_PLATFORM")
	plat = platformDetails{}
	plat.setPlatformDetails(strings.ToLower(env))

	if err := templates.ExecuteTemplate(w, "home", map[string]interface{}{
		"session_id":    sessionID(r),
		"request_id":    r.Context().Value(ctxKeyRequestID{}),
		"user_currency": currentCurrency(r),
		"show_currency": true,
		"currencies":    currencies,
		"products":      ps,
		// "cart_size":     cartSize(cart),
		"banner_color":  os.Getenv("BANNER_COLOR"), // illustrates canary deployments
		"ad":            fe.chooseAd(r.Context(), []string{}, log),
		"platform_css":  plat.css,
		"platform_name": plat.provider,
	}); err != nil {
		log.Error(err)
	}
}

func (plat *platformDetails) setPlatformDetails(env string) {
	if env == "aws" {
		plat.provider = "AWS"
		plat.css = "aws-platform"
	} else if env == "onprem" {
		plat.provider = "On-Premises"
		plat.css = "onprem-platform"
	} else if env == "azure" {
		plat.provider = "Azure"
		plat.css = "azure-platform"
	} else {
		plat.provider = "Google Cloud"
		plat.css = "gcp-platform"
	}
}

func (fe *frontendServer) productHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	id := mux.Vars(r)["id"]
	if id == "" {
		renderHTTPError(log, r, w, errors.New("product id not specified"), http.StatusBadRequest)
		return
	}
	log.WithField("id", id).WithField("currency", currentCurrency(r)).
		Debug("serving product page")

	p, err := fe.getProduct(r.Context(), id)
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve product"), http.StatusInternalServerError)
		return
	}
	currencies, err := fe.getCurrencies(r.Context())
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve currencies"), http.StatusInternalServerError)
		return
	}

	// cart, err := fe.getCart(r.Context(), sessionID(r))
	// if err != nil {
	// 	renderHTTPError(log, r, w, errors.Wrap(err, "could not retrieve cart"), http.StatusInternalServerError)
	// 	return
	// }

	price, err := fe.convertCurrency(r.Context(), p.GetPriceUsd(), currentCurrency(r))
	if err != nil {
		renderHTTPError(log, r, w, errors.Wrap(err, "failed to convert currency"), http.StatusInternalServerError)
		return
	}

	product := struct {
		Item  *pb.Product
		Price *pb.Money
	}{p, price}

	if err := templates.ExecuteTemplate(w, "product", map[string]interface{}{
		"session_id":      sessionID(r),
		"request_id":      r.Context().Value(ctxKeyRequestID{}),
		"ad":              fe.chooseAd(r.Context(), p.Categories, log),
		"user_currency":   currentCurrency(r),
		"show_currency":   true,
		"currencies":      currencies,
		"product":         product,
		"platform_css":    plat.css,
		"platform_name":   plat.provider,
	}); err != nil {
		log.Println(err)
	}
}

func (fe *frontendServer) setCurrencyHandler(w http.ResponseWriter, r *http.Request) {
	log := r.Context().Value(ctxKeyLog{}).(logrus.FieldLogger)
	cur := r.FormValue("currency_code")
	log.WithField("curr.new", cur).WithField("curr.old", currentCurrency(r)).
		Debug("setting currency")

	if cur != "" {
		http.SetCookie(w, &http.Cookie{
			Name:   cookieCurrency,
			Value:  cur,
			MaxAge: cookieMaxAge,
		})
	}
	referer := r.Header.Get("referer")
	if referer == "" {
		referer = "/"
	}
	w.Header().Set("Location", referer)
	w.WriteHeader(http.StatusFound)
}

// chooseAd queries for advertisements available and randomly chooses one, if
// available. It ignores the error retrieving the ad since it is not critical.
func (fe *frontendServer) chooseAd(ctx context.Context, ctxKeys []string, log logrus.FieldLogger) *pb.Ad {
	ads, err := fe.getAd(ctx, ctxKeys)
	if err != nil {
		log.WithField("error", err).Warn("failed to retrieve ads")
		return nil
	}
	return ads[rand.Intn(len(ads))]
}

func renderHTTPError(log logrus.FieldLogger, r *http.Request, w http.ResponseWriter, err error, code int) {
	log.WithField("error", err).Error("request error")
	errMsg := fmt.Sprintf("%+v", err)

	w.WriteHeader(code)
	if templateErr := templates.ExecuteTemplate(w, "error", map[string]interface{}{
		"session_id":  sessionID(r),
		"request_id":  r.Context().Value(ctxKeyRequestID{}),
		"error":       errMsg,
		"status_code": code,
		"status":      http.StatusText(code),
	}); templateErr != nil {
		log.Println(templateErr)
	}
}

func currentCurrency(r *http.Request) string {
	c, _ := r.Cookie(cookieCurrency)
	if c != nil {
		return c.Value
	}
	return defaultCurrency
}

func sessionID(r *http.Request) string {
	v := r.Context().Value(ctxKeySessionID{})
	if v != nil {
		return v.(string)
	}
	return ""
}


func renderMoney(money pb.Money) string {
	return fmt.Sprintf("%s %d.%02d", money.GetCurrencyCode(), money.GetUnits(), money.GetNanos()/10000000)
}
