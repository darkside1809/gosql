package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/darkside1809/gosql/cmd/app/middleware"
	"github.com/darkside1809/gosql/pkg/customers"
	"github.com/darkside1809/gosql/pkg/managers"
	"github.com/darkside1809/gosql/pkg/security"
	"github.com/gorilla/mux"
)
type Server struct {
	mux          *mux.Router
	customersSvc *customers.Service
	securitySvc  *security.Service
	managersSvc	 *managers.Service
}

const (
	GET = "GET"
	POST = "POST"
	DELETE = "DELETE"
)

func NewServer(mux *mux.Router, customersSvc	*customers.Service, securitySvc *security.Service, managersSvc *managers.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc, securitySvc: securitySvc, managersSvc: managersSvc}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
func responceByJson(w http.ResponseWriter, d interface{}) {
	data, err := json.Marshal(d)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}
// Init server with its routes
func (s *Server) Init() {
	// Authenticate customers routes by token and create prefix /api/customers
	customersAuthenticateMd := middleware.Authenticate(s.customersSvc.IDByToken)
	customersSubrouter := s.mux.PathPrefix("/api/customers").Subrouter()
	customersSubrouter.Use(customersAuthenticateMd)
	// Customers routes 
	customersSubrouter.HandleFunc("", s.handleRegisterCustomer).Methods(POST)
	customersSubrouter.HandleFunc("/token", s.handleGetCustomerToken).Methods(POST)
	customersSubrouter.HandleFunc("/products", s.handleCustomerGetProducts).Methods(GET)
	customersSubrouter.HandleFunc("/purchases", s.handleCustomerGetPurchases).Methods(GET)
	//customersSubrouter.HandleFunc("/purchases", s.handleCustomerMakePurchases).Methods(POST)

	//Authenticate managers routes by token and create prefix /api/managers
	managersAuthenticateMd := middleware.Authenticate(s.managersSvc.IDByToken)
	managersSubrouter := s.mux.PathPrefix("/api/managers").Subrouter()
	managersSubrouter.Use(managersAuthenticateMd)
	// Managers routes
	managersSubrouter.HandleFunc("", s.handleManagerRegistration).Methods(POST)
	managersSubrouter.HandleFunc("/token", s.handleManagerGetToken).Methods(POST)
	managersSubrouter.HandleFunc("/sales", s.handleManagerGetSales).Methods(GET)
	managersSubrouter.HandleFunc("/sales", s.handleManagerMakeSale).Methods(POST)
	managersSubrouter.HandleFunc("/products", s.handleManagerGetProducts).Methods(GET)
	managersSubrouter.HandleFunc("/products", s.handleManagerChangeProducts).Methods(POST)
	managersSubrouter.HandleFunc("/products/{id}", s.handleManagerRemoveProductByID).Methods(DELETE)
	managersSubrouter.HandleFunc("/customers", s.handleManagerGetCustomers).Methods(GET)
	managersSubrouter.HandleFunc("/customers", s.handleManagerChangeCustomer).Methods(POST)
	managersSubrouter.HandleFunc("/customers/{id}", s.handleManagerRemoveCustomerByID).Methods(DELETE)


	// Customer's routes without prefixes
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
	
	//s.mux.HandleFunc("/customers.getById", 		s.handleGetCustomerByID)
	//s.mux.HandleFunc("/customers.getAll",			s.handleGetAllCustomers)
	//s.mux.HandleFunc("/customers.save", 			s.handleSaveCustomer)
	//s.mux.HandleFunc("/customers.removeById", 	s.handleRemoveCustomerByID)
	//s.mux.HandleFunc("/customers.getAllActive", 	s.handleGetAllActiveCustomers)
	//s.mux.HandleFunc("/customers.blockById", 		s.handleblockCustomerByID)
	//s.mux.HandleFunc("/customers.unblockById", 	s.handleUnblockCustomerByID)
}