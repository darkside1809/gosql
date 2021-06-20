package app

import (
	"net/http"
	"encoding/json"
	"log"
	"strconv"

	"github.com/darkside1809/gosql/cmd/app/middleware"
	"github.com/darkside1809/gosql/pkg/managers"
	"github.com/darkside1809/gosql/pkg/customers"
	"github.com/gorilla/mux"
)


func (s *Server) handleManagerRegistrations(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if id == 0 {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	var item struct {
		ID 	int64		`json:"id"`
		Name 	string	`json:"name"`
		Phone string	`json:"phone"`
		Roles	[]string	`json:"roles"`
	}


	err = json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Print("Can't decode login and password")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	Admin := "ADMIN"
	administrator := s.managersSvc.IsAdmin(r.Context(), id)
	if administrator != true {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	itemManager := &managers.Manager{
		ID:	 item.ID,
		Name:  item.Name,
		Phone: item.Phone,
	}

	for _, role := range item.Roles {
		if role == Admin {
			itemManager.IsAdmin = true
			break
		}
	}

	token, err := s.managersSvc.Register(r.Context(), itemManager)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerRegistration(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if id == 0 {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	const Admin = "ADMIN"
	var registrationItem struct {
		ID    int64    `json:"id"`
		Name  string   `json:"name"`
		Phone string   `json:"phone"`
		Roles []string `json:"roles"`
	}

	err = json.NewDecoder(r.Body).Decode(&registrationItem)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	administrator := s.managersSvc.IsAdmin(r.Context(),id)
	if administrator != true {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	item := &managers.Manager{
		ID:    registrationItem.ID,
		Name:  registrationItem.Name,
		Phone: registrationItem.Phone,
	}

	for _, role := range registrationItem.Roles {
		if role == Admin {
			item.IsAdmin = true
			break
		}
	}

	token, err := s.managersSvc.Register(r.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerGetToken(w http.ResponseWriter, r *http.Request) {
	var item *managers.Manager

	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	token, err := s.managersSvc.Token(r.Context(), item.Phone, item.Password)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, map[string]interface{}{"token": token})
}

func (s *Server) handleManagerGetProducts(w http.ResponseWriter, r *http.Request) {
	items, err := s.managersSvc.Products(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, items)
}

func (s *Server) handleManagerChangeProducts(w http.ResponseWriter, r *http.Request) {
	var item *managers.Product	
	err := json.NewDecoder(r.Body).Decode(&item)
		
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	
	items, err := s.managersSvc.ChangeProducts(r.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, items)
}

func (s *Server) handleManagerGetPurchases(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	items, err := s.managersSvc.Purchases(r.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, items)
}

func (s *Server) handleManagerMakeSale(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	
	var item *managers.Sale
	err = json.NewDecoder(r.Body).Decode(&item)
	item.ManagerID = id	
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	
	items, err := s.managersSvc.MakeSale(r.Context(), item)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responceByJson(w, items)
}

func (s *Server) handleManagerGetSales(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if id == 0 {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	total, err := s.managersSvc.GetSales(r.Context(), id)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	responceByJson(w, map[string]interface{}{"manager_id": id, "total": total})
}

func (s *Server) handleManagerRemoveProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if id == 0 {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	productID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.managersSvc.RemoveProductByID(r.Context(), productID)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}

func (s *Server) handleManagerChangeCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if id == 0 {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	customer := &customers.Customer{}
	err = json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	customer, err = s.managersSvc.ChangeCustomer(r.Context(), customer)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	responceByJson(w, customer)
}

func (s *Server) handleManagerGetCustomers(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())

	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if id == 0 {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	items, err := s.managersSvc.GetCustomers(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	responceByJson(w, items)
}

func (s *Server) handleManagerRemoveCustomerByID(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.Authentication(r.Context())
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if id == 0 {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	customerID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = s.managersSvc.RemoveCustomerByID(r.Context(), customerID)
	if err != nil {
		log.Print(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}