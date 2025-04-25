package alor

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	defer goleak.VerifyTestMain(m)

	exitVal := m.Run()

	os.Exit(exitVal)
}

func TestEnqueueElement(t *testing.T) {
	t.Parallel()

	var (
		element *ChainEvent = &ChainEvent{
			Opcode: BarsOpcode,
			Guid:   "bars_guid",
		}
		//userID       int               = 1
		//skuID        int               = 1
		//count        uint              = 10
		//respc        chan handler.Item = make(chan handler.Item, 1)
		//errc         chan error        = make(chan error, 1)
		//wg           sync.WaitGroup    = sync.WaitGroup{}
		//ctx          context.Context   = minimock.AnyContext
		//itemResponse handler.Item      = handler.Item{
		//	SkuID: skuID,
		//	Name:  "Nike",
		//	Price: 45,
		//	Count: 2,
		//}
	)

	//ctrl := minimock.NewController(t)
	//productServiceMock := NewProductServiceMock(ctrl)
	//repoMock := NewRepositoryMock(ctrl)
	//lomsMock := NewLomsServiceMock(ctrl)
	//addHandler := New(repoMock, productServiceMock, lomsMock)
	//
	//repoMock.AddItemMock.When(userID, skuID, count).Then(nil)
	//lomsMock.StocksInfoMock.When(1).Then(200, nil)
	//
	//// productServiceMock.GetProductMock.When(minimock.AnyContext, skuID).Then(product, nil)
	//productServiceMock.AddWaitItemsMock.Expect(1).Return()
	//productServiceMock.DoneWaitItemMock.Return()
	//productServiceMock.CreateNewRespChanMock.When(1).Then(respc)
	//productServiceMock.CreateNewErrChanMock.Return(errc)
	//productServiceMock.CreateNewWgMock.Return(&wg)
	//productServiceMock.PutAsyncRequestToQueueMock.Set(func(data *handler.ProductRequestData) {
	//	respc <- itemResponse
	//	wg.Done()
	//})

	queue := NewChainQueue(1000)

	err := queue.Enqueue(element)

	require.NoError(t, err)
	require.Equal(t, 1, queue.GetLength())
}

func TestDequeueElement(t *testing.T) {
	t.Parallel()

	var (
		element = &ChainEvent{
			Opcode: BarsOpcode,
			Guid:   "bars_guid",
		}
	)

	queue := NewChainQueue(1000)

	err := queue.Enqueue(element)

	deqElement, err := queue.Dequeue()

	require.NoError(t, err)
	require.Equal(t, "bars_guid", deqElement.Guid)
}

func TestEnqueueSomeElements(t *testing.T) {
	t.Parallel()

	queue := NewChainQueue(1000)

	for i := range 1000 {
		element := &ChainEvent{
			Opcode: BarsOpcode,
			Guid:   GUID(fmt.Sprintf("bars_guid_%d", i+1)),
		}

		err := queue.Enqueue(element)
		require.NoError(t, err)
		require.Equal(t, i+1, queue.GetLength())
	}

	for i := range 1000 {
		element, err := queue.Dequeue()

		require.NoError(t, err)
		require.Equal(t, fmt.Sprintf("bars_guid_%d", i+1), element.Guid)
		require.Equal(t, 1000-i-1, queue.GetLength())
	}
}

func TestQueueOverFlow(t *testing.T) {
	t.Parallel()

	var wantErr = ErrQueueOverFlow

	queue := NewChainQueue(100)

	for i := range 100 {
		element := &ChainEvent{
			Opcode: BarsOpcode,
			Guid:   GUID(fmt.Sprintf("bars_guid_%d", i+1)),
		}

		err := queue.Enqueue(element)
		require.NoError(t, err)
		require.Equal(t, i+1, queue.GetLength())
	}

	err := queue.Enqueue(&ChainEvent{
		Opcode: BarsOpcode,
		Guid:   "bars_guid_101",
	})
	require.Equal(t, wantErr, err)
}

func TestQueueUnderFlow(t *testing.T) {
	t.Parallel()

	var wantErr = ErrQueueUnderFlow

	queue := NewChainQueue(100)

	element, err := queue.Dequeue()
	require.Equal(t, wantErr, err)
	require.Nil(t, element)
}
