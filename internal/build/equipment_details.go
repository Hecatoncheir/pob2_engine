package build

import "fmt"

type EquipmentDetailsQuery struct {
	Slot    string `json:"slot,omitempty"`
	ItemID  uint32 `json:"item_id,omitempty"`
	Raw     string `json:"raw,omitempty"`
	Compare bool   `json:"compare,omitempty"`
}

func (q EquipmentDetailsQuery) Validate() error {
	if (q.ItemID == 0) == (q.Raw == "") || len(q.Raw) > 65536 {
		return fmt.Errorf("нужен текст предмета либо его ID")
	}
	if q.Compare && q.Slot == "" {
		return fmt.Errorf("для сравнения нужен слот")
	}
	return nil
}
