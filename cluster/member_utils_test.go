package cluster

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func Test_KRandomMember(t *testing.T)  {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("error : %s\n", err)
			t.FailNow()
		}
	}()

	slice := make([]*Member, 6)
	for i := 0; i < 6; i ++ {
		slice[i] = &Member{
			Ip: strconv.FormatInt(int64(i), 10) + "." + strconv.FormatInt(int64(i), 10) + "." + strconv.FormatInt(int64(i), 10) + "." + strconv.FormatInt(int64(i), 10),
			Port:8080,
			Status: i,
		}
	}

	result := KRandomMember(3, slice, func(m *Member) bool {
		return m.Status % 2 == 0
	})

	assert.EqualValues(t, 3, len(result), "select member size must be 3")

	for _, e := range result {
		fmt.Printf("result : %+v\n", *e)
	}

}
