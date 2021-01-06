package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	max := 5
	rollingNumber := NewRollingNumber(1000, 10)

	for i:= 0; i < 9; i++ {
		go func() {
			sum := rollingNumber.GetRollingSum()
			fmt.Println(sum)

			if sum > max {
				fmt.Println("开始限流")
				return
			} else {
				fmt.Println("不限流")

				rollingNumber.Increment()
			}
		} ()
	}
}


type Bucket struct {
	//bucket的开始时间
	windowStart time.Time
	//结束时间
	windowEnd time.Time
	count int
}


func NewBucket(startTime time.Time, bucketSizeInMilliseconds int) *Bucket {
	return &Bucket{
		windowStart:startTime,
		windowEnd:startTime.Add(time.Duration(bucketSizeInMilliseconds) * time.Millisecond),
		count:0,
	}
}

func (b *Bucket) incr() {
	b.count++
}


func (b *Bucket) getCount() int {
	return b.count
}


type RollingNumber struct {
	//开始时间
	time time.Time
	//时间长度
	timeInMilliseconds int
	//时间长度内分成几段
	numberOfBuckets int
	//每段的时间长度
	bucketSizeInMilliseconds int

	//每段的时间和基数，环形数组
	bucketList []*Bucket
	//bucket列表容量，不变
	bucketListLength int
	//列表实际有用的长度，变化
	bucketListSize int
	//头index
	bucketTail int
	//尾index，下一个可用位
	bucketHead int

	lock sync.RWMutex
}


func NewRollingNumber(timeInMilliseconds, numberOfBuckets int) *RollingNumber {
	bucketListLength := numberOfBuckets + 1

	return &RollingNumber{
		//当前时间
		time : time.Now(),
		//1000ms
		timeInMilliseconds : timeInMilliseconds,
		//10000ms分成10个bucket
		numberOfBuckets : numberOfBuckets,
		//每个bucket 100ms
		bucketSizeInMilliseconds : timeInMilliseconds / numberOfBuckets,

		//bucket列表，多1个bucket才能滑动起来
		bucketList : make([]*Bucket, bucketListLength),
		bucketListLength : bucketListLength,
		bucketListSize : 0,
		//bucket列表头节点
		bucketHead : 0,
		//bucket列表尾节点
		bucketTail : 0,

		lock : sync.RWMutex{},
	}
}


func (rn *RollingNumber) Increment()  {
	rn.lock.Lock()
	defer rn.lock.Unlock()

	rn.GetCurrentBucket().incr()
}


//获取数组计数和
func (rn *RollingNumber) GetRollingSum() int {
	rn.lock.RLock()
	defer rn.lock.RUnlock()

	lastBucket := rn.GetCurrentBucket()
	if lastBucket == nil {
		return 0
	}

	sum := 0

	head := rn.bucketHead
	tail := rn.bucketTail
	index := head

	if head > tail {
		tail += rn.bucketListLength
	}

	for {
		if head > tail {
			break
		}

		index = head % rn.bucketListLength
		sum += rn.bucketList[index].getCount()

		head++
	}

	return sum
}


func (rn *RollingNumber) convertIndex(index int) int {
	return (index + rn.bucketHead) % rn.bucketListLength
}

//获取最后一个bucket
func (rn *RollingNumber) bucketListPeekLast() *Bucket {
	if rn.bucketListSize == 0 {
		return nil
	}

	last := rn.convertIndex(rn.bucketTail - 1)
	return rn.bucketList[last]
}

//将bucket添加到列表最后
func (rn *RollingNumber) bucketListAddLast(bucket *Bucket) {
	//直接放在tail位置
	rn.bucketList[rn.bucketTail] = bucket

	//没看懂，照抄的
	if rn.bucketListSize == rn.bucketListLength {
		rn.bucketHead = (rn.bucketHead + 1) % rn.bucketListLength
		rn.bucketTail = (rn.bucketTail + 1) % rn.bucketListLength
	} else {
		rn.bucketTail = (rn.bucketTail + 1) % rn.bucketListLength
	}

	if rn.bucketHead == 0 && rn.bucketTail == 0 {
		rn.bucketListSize = 0
	} else {
		rn.bucketListSize = (rn.bucketTail + rn.bucketListLength - rn.bucketHead) % rn.bucketListLength
	}
}


//重置bucket列表
func (rn *RollingNumber) bucketListReset() {
	for k := range rn.bucketList {
		rn.bucketList[k] = nil
	}

	rn.bucketListSize = 0
	rn.bucketTail = 0
	rn.bucketHead = 0
}


func (rn *RollingNumber) GetCurrentBucket() *Bucket {
	currentTime := time.Now()

	currentBucket := rn.bucketListPeekLast()

	//当前时间刚好落在最后一个bucket
	if currentBucket != nil && currentTime.Before(currentBucket.windowEnd) {
		return currentBucket
	}

	//bucket列表空，新建一个bucket并放到最后
	if currentBucket == nil {
		newBucket := NewBucket(currentTime, rn.bucketSizeInMilliseconds)
		rn.bucketListAddLast(newBucket)

		return newBucket
	}

	for i := 0; i < rn.numberOfBuckets; i++ {
		//获取最后一个bucket，随着for添加bucket，最后一个bucket可能会一直变化
		lastBucket := rn.bucketListPeekLast()

		//当前时间在最后一个bucket范围内
		if currentTime.Before(lastBucket.windowEnd) {
			return lastBucket
		}

		//当前时间比一个周期之后还晚，所有的bucket都失效了
		if currentTime.After(lastBucket.windowEnd.Add(time.Duration(rn.timeInMilliseconds) * time.Millisecond)) {
			//清空bucket列表
			rn.bucketListReset()

			//新建一个bucket添加到末尾
			newBucket := NewBucket(currentTime, rn.bucketSizeInMilliseconds)
			rn.bucketListAddLast(newBucket)

			return newBucket
		}

		//当前时间已经超出了窗口，但是超出的范围小于一个周期
		//新建一个bucket添加到末尾
		newBucket := NewBucket(currentTime, rn.bucketSizeInMilliseconds)
		rn.bucketListAddLast(newBucket)

		//新建后bucket后，下次循环可能就会落到对应的bucket
	}

	//没用，但是不写会提示错误
	return rn.bucketListPeekLast()
}




