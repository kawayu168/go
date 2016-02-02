// Package tests has the unit tests of package messaging.
// pubnubPresence_test.go contains the tests related to the presence requests on pubnub Api
package tests

import (
	"encoding/json"
	"fmt"
	"github.com/pubnub/go/messaging"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

// TestPresenceStart prints a message on the screen to mark the beginning of
// presence tests.
// PrintTestMessage is defined in the common.go file.
func TestPresenceStart(t *testing.T) {
	PrintTestMessage("==========Presence tests start==========")
}

const PresenceServerTimeout = 30

// TestCustomUuid subscribes to a pubnub channel using a custom uuid and then
// makes a call to the herenow method of the pubnub api. The custom id should
// be present in the response else the test fails.
func TestCustomUuid(t *testing.T) {
	assert := assert.New(t)
	uuid := "customuuid"
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, uuid)
	channel := RandomChannel()

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	unsubscribeSuccessChannel := make(chan []byte)
	unsubscribeErrorChannel := make(chan []byte)
	successGet := make(chan []byte)
	errorGet := make(chan []byte)

	go pubnubInstance.Subscribe(channel, "", successChannel, false, errorChannel)
	ExpectConnectedEvent(t, channel, "", successChannel, errorChannel)

	time.Sleep(PresenceServerTimeout * time.Second)

	go pubnubInstance.HereNow(channel, true, true, successGet, errorGet)
	select {
	case value := <-successGet:
		assert.Contains(string(value), uuid)
		var occupants struct {
			Uuids     []map[string]string
			Occupancy int
		}

		err := json.Unmarshal(value, &occupants)
		if err != nil {
			assert.Fail(err.Error())
		}

		found := false
		for _, v := range occupants.Uuids {
			if v["uuid"] == uuid {
				found = true
			}
		}

		assert.True(found)
	case err := <-errorGet:
		assert.Fail("Failed to get online users", string(err))
	case <-messaging.Timeout():
		assert.Fail("HereNow timeout")
	}

	go pubnubInstance.Unsubscribe(channel, unsubscribeSuccessChannel, unsubscribeErrorChannel)
	ExpectUnsubscribedEvent(t, channel, "", unsubscribeSuccessChannel, unsubscribeErrorChannel)

	pubnubInstance.CloseExistingConnection()
}

// TestPresence subscribes to the presence notifications on a pubnub channel and
// then subscribes to a pubnub channel. The test waits till we get a response from
// the subscribe call. The method that parses the presence response sets the global
// variable _endPresenceTestAsSuccess to true if the presence contains a join info
// on the channel and _endPresenceTestAsFailure is otherwise.
func Test0Presence(t *testing.T) {
	customUuid := "customuuid"
	testName := "Presence"
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, customUuid)
	channel := RandomChannel()

	returnPresenceChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	responseChannel := make(chan string)
	waitChannel := make(chan string)
	unsubscribeSuccessChannel := make(chan []byte)
	unsubscribeErrorChannel := make(chan []byte)

	time.Sleep(time.Duration(3) * time.Second)

	go pubnubInstance.Subscribe(channel, "", returnPresenceChannel, true, errorChannel)
	go ParseSubscribeResponseForPresence(pubnubInstance, customUuid, returnPresenceChannel, channel, testName, responseChannel)
	//go ParseResponseDummy(errorChannel)
	go ParseResponseDummyMessage(errorChannel, "aborted", responseChannel)
	go WaitForCompletion(responseChannel, waitChannel)
	ParseWaitResponse(waitChannel, t, testName)

	go pubnubInstance.Unsubscribe(channel, unsubscribeSuccessChannel, unsubscribeErrorChannel)
	ExpectUnsubscribedEvent(t, channel, "", unsubscribeSuccessChannel, unsubscribeErrorChannel)

	pubnubInstance.CloseExistingConnection()
}

// TestWhereNow subscribes to a pubnub channel and then
// makes a call to the herenow method of the pubnub api. The occupancy should
// be greater than one.
func TestWhereNow(t *testing.T) {
	assert := assert.New(t)
	uuid := "customuuid"
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, uuid)
	channel := RandomChannel()

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	unsubscribeSuccessChannel := make(chan []byte)
	unsubscribeErrorChannel := make(chan []byte)
	successGet := make(chan []byte)
	errorGet := make(chan []byte)

	go pubnubInstance.Subscribe(channel, "", successChannel, false, errorChannel)
	ExpectConnectedEvent(t, channel, "", successChannel, errorChannel)

	time.Sleep(5 * time.Second)

	go pubnubInstance.WhereNow(uuid, successGet, errorGet)
	select {
	case value := <-successGet:
		assert.Contains(string(value), channel)
	case err := <-errorGet:
		assert.Fail("Failed to get channel list", string(err))
	case <-messaging.Timeout():
		assert.Fail("WhereNow timeout")
	}

	go pubnubInstance.Unsubscribe(channel, unsubscribeSuccessChannel, unsubscribeErrorChannel)
	ExpectUnsubscribedEvent(t, channel, "", unsubscribeSuccessChannel, unsubscribeErrorChannel)

	pubnubInstance.CloseExistingConnection()
}

// TestGlobalHereNow subscribes to a pubnub channel and then
// makes a call to the herenow method of the pubnub api. The occupancy should
// be greater than one.
func TestGlobalHereNow(t *testing.T) {
	assert := assert.New(t)
	uuid := "customuuid"
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, uuid)
	channel := RandomChannel()

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	unsubscribeSuccessChannel := make(chan []byte)
	unsubscribeErrorChannel := make(chan []byte)
	successGet := make(chan []byte)
	errorGet := make(chan []byte)

	go pubnubInstance.Subscribe(channel, "", successChannel, false, errorChannel)
	ExpectConnectedEvent(t, channel, "", successChannel, errorChannel)

	time.Sleep(5 * time.Second)

	go pubnubInstance.GlobalHereNow(true, false, successGet, errorGet)
	select {
	case value := <-successGet:
		assert.Contains(string(value), channel)
	case err := <-errorGet:
		assert.Fail("Failed to get online users", string(err))
	case <-messaging.Timeout():
		assert.Fail("GlobalHereNow timeout")
	}

	go pubnubInstance.Unsubscribe(channel, unsubscribeSuccessChannel, unsubscribeErrorChannel)
	ExpectUnsubscribedEvent(t, channel, "", unsubscribeSuccessChannel, unsubscribeErrorChannel)

	pubnubInstance.CloseExistingConnection()
}

// ParseSubscribeResponseForPresence will look for the connection status in the response
// received on the go channel.
func ParseSubscribeResponseForPresence(pubnubInstance *messaging.Pubnub, customUuid string, returnChannel chan []byte, channel string, testName string, responseChannel chan string) {
	for {
		value, ok := <-returnChannel
		if !ok {
			break
		}
		//response := fmt.Sprintf("%s", value)
		//fmt.Println(response);

		if string(value) != "[]" {
			response := fmt.Sprintf("%s", value)
			message := "'" + channel + "' connected"
			messageReconn := "'" + channel + "' reconnected"
			//fmt.Println("Test3 '" + testName + "':" +response)
			if (strings.Contains(response, message)) || (strings.Contains(response, messageReconn)) {

				errorChannel2 := make(chan []byte)
				returnSubscribeChannel := make(chan []byte)
				time.Sleep(1 * time.Second)
				go pubnubInstance.Subscribe(channel, "", returnSubscribeChannel, false, errorChannel2)
				go ParseResponseDummy(returnSubscribeChannel)
				go ParseResponseDummy(errorChannel2)
			} else {
				if testName == "Presence" {
					data, _, returnedChannel, err2 := messaging.ParseJSON(value, "")

					var occupants []struct {
						Action    string
						Uuid      string
						Timestamp float64
						Occupancy int
					}

					if err2 != nil {
						responseChannel <- "Test '" + testName + "': failed. Message: 1 :" + err2.Error()
						break
					}
					//fmt.Println("Test3 '" + testName + "':" +data)
					err := json.Unmarshal([]byte(data), &occupants)
					if err != nil {
						//fmt.Println("err '" + testName + "':",err)
						responseChannel <- "Test '" + testName + "': failed. Message: 2 :" + err.Error()
						break
					} else {
						channelSubRepsonseReceived := false
						for i := 0; i < len(occupants); i++ {
							if (occupants[i].Action == "join") && occupants[i].Uuid == customUuid {
								channelSubRepsonseReceived = true
								break
							}
						}
						if !channelSubRepsonseReceived {
							responseChannel <- "Test '" + testName + "': failed. Message: err3"
							break
						}
						if channel == returnedChannel {
							responseChannel <- "Test '" + testName + "': passed."
							break
						} else {
							responseChannel <- "Test '" + testName + "': failed. Message: err4"
							break
						}
					}
				}
			}
		}
	}
}

// TestSetGetUserState subscribes to a pubnub channel and then
// makes a call to the herenow method of the pubnub api. The occupancy should
// be greater than one.
func TestSetGetUserState(t *testing.T) {
	assert := assert.New(t)
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, "")
	channel := RandomChannel()

	key := "testkey"
	val := "testval"

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	successSet := make(chan []byte)
	errorSet := make(chan []byte)
	successGet := make(chan []byte)
	errorGet := make(chan []byte)
	unsubscribeSuccessChannel := make(chan []byte)
	unsubscribeErrorChannel := make(chan []byte)

	go pubnubInstance.Subscribe(channel, "", successChannel, false, errorChannel)
	ExpectConnectedEvent(t, channel, "", successChannel, errorChannel)

	go pubnubInstance.SetUserStateKeyVal(channel, key, val, successSet, errorSet)
	select {
	case value := <-successSet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\"}", key, val)

		assert.Contains(actual, expectedSubstring)
	case err := <-errorSet:
		assert.Fail("Failed to set state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Set state timeout")
	}

	time.Sleep(10 * time.Second)

	go pubnubInstance.GetUserState(channel, successGet, errorGet)
	select {
	case value := <-successGet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\"}", key, val)

		assert.Contains(actual, expectedSubstring)
	case err := <-errorSet:
		assert.Fail("Failed to get state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Get state timeout")
	}

	go pubnubInstance.Unsubscribe(channel, unsubscribeSuccessChannel, unsubscribeErrorChannel)
	ExpectUnsubscribedEvent(t, channel, "", unsubscribeSuccessChannel, unsubscribeErrorChannel)

	pubnubInstance.CloseExistingConnection()
}

func TestSetUserStateHereNow(t *testing.T) {
	assert := assert.New(t)
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, "")
	channel := RandomChannel()

	key := "testkey"
	val := "testval"

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	successSet := make(chan []byte)
	errorSet := make(chan []byte)
	successGet := make(chan []byte)
	errorGet := make(chan []byte)
	unsubscribeSuccessChannel := make(chan []byte)
	unsubscribeErrorChannel := make(chan []byte)

	go pubnubInstance.Subscribe(channel, "", successChannel, false, errorChannel)
	ExpectConnectedEvent(t, channel, "", successChannel, errorChannel)

	go pubnubInstance.SetUserStateKeyVal(channel, key, val, successSet, errorSet)
	select {
	case value := <-successSet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\"}", key, val)

		assert.Contains(actual, expectedSubstring)
	case err := <-errorSet:
		assert.Fail("Failed to set state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Set state timeout")
	}

	time.Sleep(10 * time.Second)

	go pubnubInstance.HereNow(channel, true, true, successGet, errorGet)
	select {
	case value := <-successGet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\"}", key, val)

		assert.Contains(actual, expectedSubstring)
		assert.Contains(actual, pubnubInstance.GetUUID())
	case err := <-errorSet:
		assert.Fail("Failed to get state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Get state timeout")
	}

	go pubnubInstance.Unsubscribe(channel, unsubscribeSuccessChannel, unsubscribeErrorChannel)
	ExpectUnsubscribedEvent(t, channel, "", unsubscribeSuccessChannel, unsubscribeErrorChannel)

	pubnubInstance.CloseExistingConnection()
}

func TestSetUserStateGlobalHereNow(t *testing.T) {
	assert := assert.New(t)
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, "")
	channel := RandomChannel()

	key := "testkey"
	val := "testval"

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	successSet := make(chan []byte)
	errorSet := make(chan []byte)
	successGet := make(chan []byte)
	errorGet := make(chan []byte)
	unsubscribeSuccessChannel := make(chan []byte)
	unsubscribeErrorChannel := make(chan []byte)

	go pubnubInstance.Subscribe(channel, "", successChannel, false, errorChannel)
	ExpectConnectedEvent(t, channel, "", successChannel, errorChannel)

	go pubnubInstance.SetUserStateKeyVal(channel, key, val, successSet, errorSet)
	select {
	case value := <-successSet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\"}", key, val)

		assert.Contains(actual, expectedSubstring)
	case err := <-errorSet:
		assert.Fail("Failed to set state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Set state timeout")
	}

	time.Sleep(10 * time.Second)

	go pubnubInstance.GlobalHereNow(true, true, successGet, errorGet)
	select {
	case value := <-successGet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\"}", key, val)

		assert.Contains(actual, expectedSubstring)
		assert.Contains(actual, pubnubInstance.GetUUID())
	case err := <-errorSet:
		assert.Fail("Failed to get state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Get state timeout")
	}

	go pubnubInstance.Unsubscribe(channel, unsubscribeSuccessChannel, unsubscribeErrorChannel)
	ExpectUnsubscribedEvent(t, channel, "", unsubscribeSuccessChannel, unsubscribeErrorChannel)

	pubnubInstance.CloseExistingConnection()
}

func TestSetUserStateJSON(t *testing.T) {
	assert := assert.New(t)

	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, "")

	channel := RandomChannel()

	key1 := "testkey"
	val1 := "testval"
	key2 := "testkey2"
	val2 := "testval2"

	successSet := make(chan []byte)
	errorSet := make(chan []byte)

	jsonString := fmt.Sprintf("{\"%s\": \"%s\",\"%s\": \"%s\"}", key1, val1, key2, val2)

	go pubnubInstance.SetUserStateJSON(channel, jsonString, successSet, errorSet)
	select {
	case value := <-successSet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\", \"%s\": \"%s\"}", key2, val2, key1, val1)
		assert.Contains(actual, expectedSubstring)
	case err := <-errorSet:
		assert.Fail("Failed to set state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Set state timeout")
	}

	time.Sleep(10 * time.Second)

	go pubnubInstance.SetUserStateKeyVal(channel, key2, "", successSet, errorSet)
	select {
	case value := <-successSet:
		actual := string(value)
		expectedSubstring := fmt.Sprintf("{\"%s\": \"%s\"}", key1, val1)
		assert.Contains(actual, expectedSubstring)
	case err := <-errorSet:
		assert.Fail("Failed to set state", string(err))
	case <-messaging.Timeout():
		assert.Fail("Set state timeout")
	}
}

// TestPresenceEnd prints a message on the screen to mark the end of
// presence tests.
// PrintTestMessage is defined in the common.go file.
func TestPresenceEnd(t *testing.T) {
	PrintTestMessage("==========Presence tests end==========")
}
