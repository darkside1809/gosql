package app


import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"errors"
	"github.com/gorilla/mux"
	"strconv"

	"github.com/darkside1809/gosql/cmd/app/middleware"
	"github.com/darkside1809/gosql/pkg/customers"
	"github.com/darkside1809/gosql/pkg/security"

)

// Customers handlers
func (s *Server) handleGetCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print("Cant parse id")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	item, err := s.customersSvc.ByID(r.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, item)
}

func (s *Server) handleGetAllCustomers(w http.ResponseWriter, r *http.Request) {
	items, err := s.customersSvc.All(r.Context())
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	responceByJson(w, items)
}

func (s *Server) handleGetAllActiveCustomers(w http.ResponseWriter, r *http.Request) {
	items, err := s.customersSvc.AllActive(r.Context())
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, items)
}

func (s *Server) handleSaveCustomer(w http.ResponseWriter, r *http.Request) {
	var customer *customers.Customer
	var item *customers.Auth
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Print("Can't Decode customer")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Print(item)
	customer, err = s.customersSvc.Save(r.Context(), customer)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, customer)
}

func (s *Server) handleRemoveCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print("Missing id")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	customer, err := s.customersSvc.RemoveByID(r.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, customer)
}

func (s *Server) handleblockCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print("Cant parse id")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	customer, err := s.customersSvc.BlockAndUnblockByID(r.Context(), id, false)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, customer)
}

func (s *Server) handleUnblockCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print("Cant parse id")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	customer, err := s.customersSvc.BlockAndUnblockByID(r.Context(), id, true)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, customer)
}

func (s *Server) handleGetCustomerToken(w http.ResponseWriter, r *http.Request) {
	var auth *security.Auth
	var tok security.Token
	err := json.NewDecoder(r.Body).Decode(&auth)
	fmt.Print(auth)
	if err != nil {
		log.Print("Can't Decode login and password")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Print("Login: ", auth.Login, "Password: ", auth.Password)

	token, err := s.customersSvc.Token(r.Context(), auth.Login, auth.Password)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tok.Token = token
	responceByJson(w, tok)
}

func (s *Server) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	var fail security.ResponceFail
	var ok security.ResponceOk
	var token security.Token
	var data []byte
	code := 200

	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		log.Print("Can't Decode token")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	id, er := s.securitySvc.AuthenticateCustomer(r.Context(), token.Token)

	if er == security.ErrNoSuchUser {
		code = 404
		fail.Status = "fail"
		fail.Reason = "not found"
	} else if er == security.ErrExpired {
		code = 400
		fail.Status = "fail"
		fail.Reason = "expired"
	} else if er == nil {
		log.Print(id)
		ok.Status = "ok"
		ok.CustomerID = id
	} else {
		log.Print("err", er)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if code != 200 {
		w.WriteHeader(code)

		data, err = json.Marshal(fail)
		if err != nil {
			log.Print(err)
		}
	} else {
		data, err = json.Marshal(ok)
		if err != nil {
			log.Print(err)
		}
	}
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	return
}

func (s *Server) handleRegisterCustomer(w http.ResponseWriter, r *http.Request) {
	var item *customers.Registration

	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Print("Can't Decode login and password")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	saved, err := s.customersSvc.Register(r.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, saved)
}

func (s *Server) handleCustomerGetToken(w http.ResponseWriter, r *http.Request) {
	var item *customers.Auth
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Print("Can't Decode login and password")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := s.customersSvc.Token(r.Context(), item.Login, item.Password)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, map[string]interface{}{"token": token})
}

func (s *Server) handleCustomerGetProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.customersSvc.Products(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, items)
}

func (s *Server) handleCustomerGetPurchases(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items, err := s.customersSvc.Purchases(r.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, items)
}