package compensations_test

import (
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/core/compensations"
	"github.com/stretchr/testify/assert"
)

func TestCompensations_Run_ExecutesInReverseOrder(t *testing.T) {
	var order []int
	comp := compensations.New()
	comp.Push(func() { order = append(order, 1) })
	comp.Push(func() { order = append(order, 2) })
	comp.Push(func() { order = append(order, 3) })
	comp.Run()
	assert.Equal(t, []int{3, 2, 1}, order)
}

func TestCompensations_Pop_RemovesLastCompensation(t *testing.T) {
	var order []int
	comp := compensations.New()
	comp.Push(func() { order = append(order, 1) })
	comp.Push(func() { order = append(order, 2) })
	comp.Pop()
	comp.Run()
	assert.Equal(t, []int{1}, order)
}

func TestCompensations_Pop_OnEmptyStack_IsNoop(t *testing.T) {
	comp := compensations.New()
	comp.Pop()
	comp.Run()
}

func TestCompensations_Run_OnEmptyStack_IsNoop(t *testing.T) {
	comp := compensations.New()
	comp.Run()
}

func TestCompensations_Push_AfterRun_IsNoop(t *testing.T) {
	var calls int
	comp := compensations.New()
	comp.Run()
	comp.Push(func() { calls++ })
	assert.Equal(t, 0, calls)
}

func TestCompensations_Pop_AfterRun_IsNoop(t *testing.T) {
	var order []int
	comp := compensations.New()
	comp.Push(func() { order = append(order, 1) })
	comp.Run()
	comp.Pop()
	assert.Equal(t, []int{1}, order)
}

func TestCompensations_Run_CalledTwice_ExecutesOnce(t *testing.T) {
	var calls int
	comp := compensations.New()
	comp.Push(func() { calls++ })
	comp.Run()
	comp.Run()
	assert.Equal(t, 1, calls)
}
