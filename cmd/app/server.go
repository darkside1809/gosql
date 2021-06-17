package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"errors"
	"fmt"
	//"github.com/darkside1809/gosql/cmd/app/middleware"
	"github.com/darkside1809/gosql/pkg/customers"
	"github.com/darkside1809/gosql/pkg/security"
	"github.com/gorilla/mux"
)
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	securitySvc  *security.Service
}

const (
	GET = "GET"
	POST = "POST"
	DELETE = "DELETE"
)

func NewServer(mux *mux.Router, customersSvc	*customers.Service, securitySvc *security.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc, securitySvc: securitySvc}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}


func (s *Server) Init() {
	s.mux.HandleFunc("/api/customers/token", s.handleGetCustomerToken).Methods(POST)
	s.mux.HandleFunc("/api/customers/token/validate", s.handleValidateToken).Methods(POST)
	s.mux.HandleFunc("/api/customers", s.handleSaveCustomer).Methods(POST)

	s.mux.HandleFunc("/customers", 			s.handleGetAllCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/active", 	s.handleGetAllActiveCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/{id}", 	s.handleGetCustomerByID).Methods(GET)
	s.mux.HandleFunc("/customers", 			s.handleSaveCustomer).Methods(POST)
	s.mux.HandleFunc("/customers/{id}",		 s.handleRemoveCustomerByID).Methods(DELETE)
	s.mux.HandleFunc("/customers/{id}/block", s.handleblockCustomerByID).Methods(POST)
	s.mux.HandleFunc("/customers/{id}/block", s.handleUnblockCustomerByID).Methods(DELETE)

	// s.mux.Use(middleware.Basic(s.securitySvc.Auth))
	// s.mux.Use(middleware.Logger)
	// customerAuth := middleware.Authenticate(s.customersSvc.CustomerIDByToken)
	// s.mux.Use(customerAuth)

	//s.mux.HandleFunc("/customers.getById", 		s.handleGetCustomerByID)		//GetById
	//s.mux.HandleFunc("/customers.getAll",			s.handleGetAllCustomers)		//GetAll
	//s.mux.HandleFunc("/customers.save", 			s.handleSaveCustomer)			//SaveCustomer
	//s.mux.HandleFunc("/customers.removeById", 	s.handleRemoveCustomerByID)	//removeById
	//s.mux.HandleFunc("/customers.getAllActive", 	s.handleGetAllActiveCustomers)//GetAllActive
	//s.mux.HandleFunc("/customers.blockById", 		s.handleblockCustomerByID)		//blockById
	//s.mux.HandleFunc("/customers.unblockById", 	s.handleUnblockCustomerByID)	//unblockById
}

func (s *Server) handleGetCustomerByID(writer http.ResponseWriter, request *http.Request) {
	idParam, ok := mux.Vars(request)["id"]
	if !ok {
		log.Print("Cant parse id")
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	item, err := s.customersSvc.ByID(request.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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

	data, err := json.Marshal(items)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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

	data, err := json.Marshal(items)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleSaveCustomer(w http.ResponseWriter, r *http.Request) {
	var customer *customers.Customer
	var item *customers.CustomerAuth
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

	data, err := json.Marshal(customer)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	data, err := json.Marshal(customer)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	data, err := json.Marshal(customer)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	data, err := json.Marshal(customer)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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

	token, err := s.customersSvc.TokenForCustomer(r.Context(), auth.Login, auth.Password)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tok.Token = token
	data, err := json.Marshal(tok)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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