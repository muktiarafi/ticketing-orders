package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	common "github.com/muktiarafi/ticketing-common"
	"github.com/muktiarafi/ticketing-orders/internal/entity"
	"github.com/muktiarafi/ticketing-orders/internal/model"
)

func TestOrderHandlerCreate(t *testing.T) {
	user := &common.UserPayload{1, "bambank@gmail.com"}
	cookie := signIn(user)

	t.Run("create order normally", func(t *testing.T) {
		ticket := &entity.Ticket{
			ID:    1,
			Title: "ticket",
			Price: 12,
		}
		newTicket, err := ticketRepo.Insert(ticket)
		if err != nil {
			t.Error(err)
		}
		orderDTO := model.OrderDTO{
			TicketID: newTicket.ID,
		}
		orderDTOJSON, _ := json.Marshal(orderDTO)

		request := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(orderDTOJSON))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusCreated, response.Code)

		responseBody, _ := ioutil.ReadAll(response.Body)
		apiResponse := struct {
			Data *entity.Order `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)

		if apiResponse.Data.Status != "CREATED" {
			t.Errorf("expecting status to be 'CREATED' but got %q instead", apiResponse.Data.Status)
		}
	})

	t.Run("create order with nonexistent ticket", func(t *testing.T) {
		orderDTO := model.OrderDTO{
			TicketID: 9999991,
		}
		orderDTOJSON, _ := json.Marshal(orderDTO)

		request := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(orderDTOJSON))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusNotFound, response.Code)
	})
}

func TestOrderHandlerGetAll(t *testing.T) {
	user := &common.UserPayload{2, "bambank@gmail.com"}
	cookie := signIn(user)

	t.Run("get all order when not ordering ticket", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusOK, response.Code)

		responseBody, _ := ioutil.ReadAll(response.Body)
		apiResponse := struct {
			Data []*entity.Order `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)

		got := len(apiResponse.Data)
		want := 0
		if got != want {
			t.Errorf("expecting the number of orders to be %d but got %d instead", want, got)
		}
	})

	t.Run("get all orders after ordering some tickets", func(t *testing.T) {
		tickets := []*entity.Ticket{
			{
				ID:    2,
				Title: "a",
				Price: 2,
			},
			{
				ID:    3,
				Title: "b",
				Price: 2,
			},
			{
				ID:    4,
				Title: "c",
				Price: 2,
			},
		}

		for _, v := range tickets {
			ticketRepo.Insert(v)

			orderDTO := &model.OrderDTO{
				TicketID: v.ID,
			}
			orderDTOJSON, _ := json.Marshal(orderDTO)
			request := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(orderDTOJSON))
			request.Header.Set("Content-Type", "application/json")
			request.AddCookie(cookie)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			assertResponseCode(t, http.StatusCreated, response.Code)
		}

		request := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusOK, response.Code)

		responseBody, _ := ioutil.ReadAll(response.Body)
		apiResponse := struct {
			Data []*entity.Order `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)

		got := len(apiResponse.Data)
		want := 3
		if got != want {
			t.Errorf("expecting the number of orders to be %d but got %d instead", want, got)
		}
	})
}

func TestOrderHandlerShow(t *testing.T) {
	user := &common.UserPayload{3, "bambank@gmail.com"}
	cookie := signIn(user)

	t.Run("show nonexistent order", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/orders/8834", nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusNotFound, response.Code)
	})

	t.Run("show order after creating one", func(t *testing.T) {
		ticket := &entity.Ticket{
			ID:    6,
			Title: "ticket",
			Price: 12,
		}
		newTicket, err := ticketRepo.Insert(ticket)
		if err != nil {
			t.Error(err)
		}
		orderDTO := model.OrderDTO{
			TicketID: newTicket.ID,
		}
		orderDTOJSON, _ := json.Marshal(orderDTO)

		request := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(orderDTOJSON))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)
		assertResponseCode(t, http.StatusCreated, response.Code)

		responseBody, _ := ioutil.ReadAll(response.Body)
		apiResponse := struct {
			Data *entity.Order `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)

		orderID := apiResponse.Data.ID
		request = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/orders/%d", orderID), nil)
		request.AddCookie(cookie)
		response = httptest.NewRecorder()

		router.ServeHTTP(response, request)
		assertResponseCode(t, http.StatusOK, response.Code)

		responseBody, _ = ioutil.ReadAll(response.Body)
		apiResponse = struct {
			Data *entity.Order `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)

		got := apiResponse.Data.ID
		if got != orderID {
			t.Errorf("expecting order id to be %d but got %d instead", orderID, got)
		}
	})
}

func TestOrderHandlerUpdate(t *testing.T) {
	user := &common.UserPayload{4, "bambank@gmail.com"}
	cookie := signIn(user)

	t.Run("update after creating ticket", func(t *testing.T) {
		ticket := &entity.Ticket{
			ID:    9,
			Title: "ticket",
			Price: 12,
		}
		newTicket, err := ticketRepo.Insert(ticket)
		if err != nil {
			t.Error(err)
		}
		orderDTO := model.OrderDTO{
			TicketID: newTicket.ID,
		}
		orderDTOJSON, _ := json.Marshal(orderDTO)

		request := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(orderDTOJSON))
		request.Header.Set("Content-Type", "application/json")
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		assertResponseCode(t, http.StatusCreated, response.Code)
		responseBody, _ := ioutil.ReadAll(response.Body)
		apiResponse := struct {
			Data *entity.Order `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)

		request = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/orders/%d", apiResponse.Data.ID), nil)
		request.AddCookie(cookie)
		response = httptest.NewRecorder()

		router.ServeHTTP(response, request)
		assertResponseCode(t, http.StatusOK, response.Code)

		order := apiResponse.Data
		responseBody, _ = ioutil.ReadAll(response.Body)
		apiResponse = struct {
			Data *entity.Order `json:"data"`
		}{}
		json.Unmarshal(responseBody, &apiResponse)
		updatedOrder := apiResponse.Data

		if order.ID != updatedOrder.ID {
			t.Errorf("expecting order id to be %d but got %d instead", order.ID, updatedOrder.ID)
		}

		if updatedOrder.Status != "CANCELLED" {
			t.Errorf("expecting status to be 'CANCELLED' but got %q instead", updatedOrder.Status)
		}

		if updatedOrder.Version != order.Version+1 {
			t.Errorf("expecting version to be incremented but got %d instead", updatedOrder.Version)
		}
	})

	t.Run("update nonexistent order", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/orders/1312312", nil)
		request.AddCookie(cookie)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)
		assertResponseCode(t, http.StatusNotFound, response.Code)
	})
}
