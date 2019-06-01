package utils

import (
	"math/rand"
	"time"

	defs "../defs"
)

func pickWinner() {
	time.Sleep(30 * time.Second)
	defs.Mutex.Lock()
	temp := defs.TempBlocks
	defs.Mutex.Unlock()

	lotteryPool := []string{}
	if len(temp) > 0 {

		// slightly modified traditional proof of stake algorithm
		// from all validators who submitted a block, weight them by the number of staked tokens
		// in traditional proof of stake, validators can participate without submitting a block to be forged
	OUTER:
		for _, block := range temp {
			// if already in lottery pool, skip
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			// lock list of validators to prevent data race
			defs.Mutex.Lock()
			setValidators := defs.Validators
			defs.Mutex.Unlock()

			k, ok := setValidators[block.Validator]
			if ok {
				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}
		}

		// randomly pick winner from lottery pool
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)
		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

		// add block of winner to blockchain and let all the other nodes know
		for _, block := range temp {
			if block.Validator == lotteryWinner {
				defs.Mutex.Lock()
				defs.Blockchain = append(defs.Blockchain, block)
				defs.Mutex.Unlock()
				break
			}
		}
	}

	defs.Mutex.Lock()
	defs.TempBlocks = []defs.Block{}
	defs.Mutex.Unlock()
}
