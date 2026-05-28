package service

import (
	"item-repository/model"
)

type InventoryRepository interface {
	GetItemByName(name string) (model.Item, error)
	GetItemByID(id uint) (model.Item, error)
	GetItems() ([]model.Item, error)
	ReduceItemNum(name string, num int) error
	RepositoryInit() error
	AddProcessedOrder(orderID uint) error
	IsOrderProcessed(orderID uint) (bool, error)
	ProcessOrder(orderID uint, itemname string, nums int) error
}

type ItemService struct {
	Itemrepo InventoryRepository
}

type ItemDTO struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
}

func (s *ItemService) ConvertToDTO(item model.Item) ItemDTO {
	return ItemDTO{
		ID:    item.ID,
		Name:  item.Name,
		Stock: item.AvailableStock,
	}
}

func NewItemService(itemRepo InventoryRepository) *ItemService {
	return &ItemService{
		Itemrepo: itemRepo,
	}
}

func (s *ItemService) GetItemByName(name string) (ItemDTO, error) {
	item, err := s.Itemrepo.GetItemByName(name)
	if err != nil {
		return ItemDTO{}, err
	}
	return s.ConvertToDTO(item), nil
}

func (s *ItemService) GetItemByID(id uint) (ItemDTO, error) {
	item, err := s.Itemrepo.GetItemByID(id)
	if err != nil {
		return ItemDTO{}, err
	}
	return s.ConvertToDTO(item), nil
}

func (s *ItemService) GetItems() ([]ItemDTO, error) {
	items, err := s.Itemrepo.GetItems()
	if err != nil {
		return nil, err
	}
	dtos := make([]ItemDTO, len(items))
	for i, item := range items {
		dtos[i] = s.ConvertToDTO(item)
	}
	return dtos, nil
}

func (s *ItemService) ProcessOrder(orderID uint, itemname string, nums int) error {
	return s.Itemrepo.ProcessOrder(orderID, itemname, nums)
}
