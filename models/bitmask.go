package models

import (
	"fmt"
)

type BitMask []uint64

// EnsureCapacity ensures the bitmap has enough capacity for the given bit index
func (b *BitMask) EnsureCapacity(bitIndex uint) {
	requiredLen := (bitIndex + 64) / 64
	if requiredLen > uint(len(*b)) {
		// Extend the slice to accommodate the new bit index
		newBits := make([]uint64, requiredLen)
		copy(newBits, *b)
		*b = newBits
	}
}

// Set sets the bit at the given index
func (b *BitMask) Set(index uint, value bool) {
	b.EnsureCapacity(index)
	if value {
		(*b)[index/64] |= 1 << (index % 64) // Set bit to 1
	} else {
		(*b)[index/64] &^= 1 << (index % 64) // Set bit to 0
	}
}

// Clear clears the bit at the given index
func (b *BitMask) Clear(index uint) {
	b.EnsureCapacity(index)
	(*b)[index/64] &^= 1 << (index % 64)
}

// IsSet checks if the bit at the given index is set
func (b *BitMask) IsSet(index int) bool {
	if index/64 >= len(*b) {
		return false // Out of bounds
	}
	return (*b)[index/64]&(1<<(index%64)) != 0
}

// And performs a bitwise AND operation with another bitmap, padding the smaller one with zeroes
func (b *BitMask) And(other *BitMask) *BitMask {
	// Determine the length of the longer BitMask
	minLen := len(*b)
	if len(*other) < minLen {
		minLen = len(*other)
	}

	// Perform the AND operation
	result := make(BitMask, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = (*other)[i] & (*b)[i]
	}

	// Append zeroes for the remaining length of the longer BitMask
	if len(*b) > len(*other) {
		result = append(result, make([]uint64, len(*b)-len(*other))...)
	} else if len(*other) > len(*b) {
		result = append(result, make([]uint64, len(*other)-len(*b))...)
	}

	return &result
}

// Or performs a bitwise OR operation with another bitmap, padding the smaller one with zeroes
func (b *BitMask) Or(other *BitMask) *BitMask {
	// Determine the length of the longer BitMask
	minLen := len(*b)
	if len(*other) < minLen {
		minLen = len(*other)
	}

	// Perform the OR operation
	result := make(BitMask, minLen)
	for i := 0; i < minLen; i++ {
		result[i] = (*other)[i] | (*b)[i]
	}

	if len(*b) > len(*other) {
		// Append the remaining bits from the longer BitMask
		result = append(result, (*b)[minLen:]...)
	} else if len(*other) > len(*b) {
		// Append the remaining bits from the longer BitMask
		result = append(result, (*other)[minLen:]...)
	}

	return &result
}

func (b *BitMask) Not() *BitMask {
	result := make(BitMask, len(*b))
	for i := 0; i < len(*b); i++ {
		result[i] = ^(*b)[i]
	}
	return &result
}

// Print displays the bitmap as a binary string for debugging, MSB first
func (b *BitMask) Print() {
	for i := 0; i < len(*b); i++ {
		segment := (*b)[i]
		var reversed uint64
		for j := 0; j < 64; j++ {
			reversed |= ((segment >> j) & 1) << (63 - j)
		}
		fmt.Printf("%064b ", reversed)
	}
	fmt.Println()
}
