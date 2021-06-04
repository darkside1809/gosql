package app

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/darkside1809/gosql/pkg/customers"
)

type Server struct {
	mux				*http.ServeMux
	customersSvc	*customers.Service
}

func NewServer(mux *http.ServeMux, customersSvc	*customers.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
func (s *Server) Init() {
	s.mux.HandleFunc("/customers.getById", 		s.handleGetCustomerByID)		//GetById
	s.mux.HandleFunc("/customers.getAll",			s.handleGetAllCustomers)		//GetAll
	s.mux.HandleFunc("/customers.getAllActive", 	s.handleGetAllActiveCustomers)//GetAllActive
	s.mux.HandleFunc("/customers.save", 			s.handleSaveCustomer)			//SaveCustomer
	s.mux.HandleFunc("/customers.removeById", 	s.handleRemoveCustomerByID)	//removeById
	s.mux.HandleFunc("/customers.blockById", 		s.handleblockCustomerByID)		//blockById
	s.mux.HandleFunc("/customers.unblockById", 	s.handleUnblockCustomerByID)	//unblockById
}

func (s *Server) handleGetCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
	}

	item, err := s.customersSvc.ByID(r.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := s.customersSvc.All(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Print(err)
	}
	data, err := json.Marshal(customers)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetAllActiveCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := s.customersSvc.AllActive(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Print(err)
	}
	data, err := json.Marshal(customers)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleSaveCustomer(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	phone := r.URL.Query().Get("phone")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	if name == "" && phone == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)		
		return
	}

	 item := &customers.Customer{
		 ID: id,
		 Name: name,
		 Phone: phone,
	 }
	 customer, err := s.customersSvc.Save(r.Context(), item)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Print(err)
	}
	data, err := json.Marshal(customer)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleRemoveCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	item, err := s.customersSvc.RemoveByID(r.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Print(err)
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleblockCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	item, err := s.customersSvc.BlockAndUnblockByID(r.Context(), id, false)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Print(err)
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleUnblockCustomerByID(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	item, err := s.customersSvc.BlockAndUnblockByID(r.Context(), id, true)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Print(err)
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}
