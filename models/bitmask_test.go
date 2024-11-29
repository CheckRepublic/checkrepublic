package models

import (
	"testing"
)

func TestBitMask_Set(t *testing.T) {
	b := BitMask{}
	b.Set(0, true)
	if !b.IsSet(0) {
		t.Errorf("Expected bit 0 to be set")
	}
}

func TestBitMask_Set_Oversize(t *testing.T) {
	b := BitMask{}
	b.Set(64, true)
	if !b.IsSet(64) {
		t.Errorf("Expected bit 0 to be set")
	}
	if (len(b)) != 2 {
		t.Errorf("Expected length to be 2")
	}
	if b[0] != 0 {
		t.Errorf("Expected first element to be 0")
	}
}

func TestBitMask_And_True(t *testing.T) {
	b1 := BitMask{}
	b1.Set(0, true)
	b2 := BitMask{}
	b2.Set(0, true)
	result := b1.And(&b2)
	if result.IsSet(0) {
		return
	}
	t.Errorf("Expected bit 0 to be set")
}

func TestBitMask_And_False(t *testing.T) {
	b1 := BitMask{}
	b1.Set(0, true)
	b2 := BitMask{}
	result := b1.And(&b2)
	if !result.IsSet(0) {
		return
	}
	t.Errorf("Expected bit 0 to be unset")
}

func TestBitMask_And_DifferentLengths(t *testing.T) {
	b1 := BitMask{}
	b1.Set(0, true)
	b2 := BitMask{}
	b2.Set(0, true)
	b2.Set(64, true)
	result := b1.And(&b2)
	if result.IsSet(0) && !result.IsSet(64) && len(*result) == 2 {
		return
	}
	t.Errorf("Expected bit 0 to be unset")
}

func TestBitMask_Or_True(t *testing.T) {
	b1 := BitMask{}
	b1.Set(0, true)
	b2 := BitMask{}
	result := b1.Or(&b2)
	if result.IsSet(0) {
		return
	}
	t.Errorf("Expected bit 0 to be set")
}

func TestBitMask_Or_False(t *testing.T) {
	b1 := BitMask{}
	b2 := BitMask{}
	result := b1.Or(&b2)
	if !result.IsSet(0) {
		return
	}
	t.Errorf("Expected bit 0 to be unset")
}

func TestBitMask_Or_DifferentLengths(t *testing.T) {
	b1 := BitMask{}
	b1.Set(0, true)
	b2 := BitMask{}
	b2.Set(64, true)
	result := b1.Or(&b2)
	if result.IsSet(0) && result.IsSet(64) && len(*result) == 2 {
		return
	}
	t.Errorf("Expected bit 0 and 65 to be set")
}

func TestBitMask_Not(t *testing.T) {
	b1 := BitMask{}
	b1.Set(0, true)
	result := b1.Not()
	if !result.IsSet(0) {
		return
	}
	t.Errorf("Expected bit 0 to be unset")
}

func TestBitMask_Not_And(t *testing.T) {
	b1 := BitMask{}
	b1.Set(0, true)
	b2 := BitMask{}
	b2.Set(0, true)
	result := b1.Not().And(&b2)
	if !result.IsSet(0) {
		return
	}
	t.Errorf("Expected bit 0 to be unset")
}
