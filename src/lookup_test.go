package main

import (
	"testing"
	"sync"
	"sort"
)

func TestGetResponse(t *testing.T) {
	receiveCh := make(chan LookupResponse)
	defer close(receiveCh)

	sendCh := make(chan Contact)
	defer close(sendCh)


	id := NewKademliaID("1111111100000000000000000000000000000000")

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact1.CalcDistance(id)
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact2.CalcDistance(id)
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")
	contact3.CalcDistance(id)

	lookup := Lookup{
		id: id,
		notCalled: []Contact{contact1},
		called: append([]Contact{contact2}),
		wg: sync.WaitGroup{},
		sendCh: sendCh,
		receiveCh: receiveCh,
		closestHasNotValue: contact1,
		mutex: sync.RWMutex{}}


	go func(){
		contacts := LookupResponse{[]Contact{contact2, contact3}, contact2, false}
		receiveCh <- contacts
	}()

	newContacts := lookup.getResponse()

	checkContacts(t, newContacts, []Contact{contact3})
	checkContacts(t, lookup.called, []Contact{contact2})


	expContacts := []Contact{contact1, contact3}
	sort.Sort(ByDistance(expContacts))
	checkContacts(t, lookup.notCalled, expContacts)
}

func TestCallContact(t *testing.T) {
	receiveCh := make(chan LookupResponse)
	defer close(receiveCh)

	sendCh := make(chan Contact)
	defer close(sendCh)


	id := NewKademliaID("1111111100000000000000000000000000000000")

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact1.CalcDistance(id)
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact2.CalcDistance(id)
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")
	contact3.CalcDistance(id)

	lookup := Lookup{
		id: id,
		notCalled: []Contact{contact1},
		called: append([]Contact{contact2}),
		wg: sync.WaitGroup{},
		sendCh: sendCh,
		receiveCh: receiveCh,
		closestHasNotValue: contact1,
		mutex: sync.RWMutex{}}

	lookup.wg.Add(1)
	go func(){
		defer lookup.wg.Done()
		contact := <-sendCh
		if !contact.ID.Equals(contact1.ID) {
			t.Error("Expected contact ", contact1, " got ", contact)
		}
	}()

	lookup.callContact()

	expContacts := []Contact{contact1, contact2}
	sort.Sort(ByDistance(expContacts))

	checkContacts(t, lookup.called, expContacts)
	checkContacts(t, lookup.notCalled, []Contact{})

	lookup.wg.Wait()
}

func TestLastEffort(t *testing.T) {
	wg := sync.WaitGroup{}

	receiveCh := make(chan LookupResponse)
	defer close(receiveCh)

	sendCh := make(chan Contact)
	defer close(sendCh)


	id := NewKademliaID("1111111100000000000000000000000000000000")

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact1.CalcDistance(id)
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact2.CalcDistance(id)
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")
	contact3.CalcDistance(id)

	lookup := Lookup{
		id: id,
		notCalled: []Contact{contact1},
		called: append([]Contact{contact2}),
		wg: sync.WaitGroup{},
		sendCh: sendCh,
		receiveCh: receiveCh,
		closestHasNotValue: contact1,
		mutex: sync.RWMutex{}}

	wg.Add(1)
	go func(){
		defer wg.Done()
		contact := <-sendCh
		if !contact.ID.Equals(contact1.ID) {
			t.Error("Expected contact ", contact1, " got ", contact)
		}
		receiveCh <- LookupResponse{[]Contact{}, contact, false}
	}()

	lookup.lastEffort()

	expContacts := []Contact{contact1, contact2}
	sort.Sort(ByDistance(expContacts))

	checkContacts(t, lookup.called, expContacts)
	checkContacts(t, lookup.notCalled, []Contact{})

	wg.Wait()
}

func TestRunLookupLessThenK(t *testing.T) {
	receiveCh := make(chan LookupResponse)
	defer close(receiveCh)

	sendCh := make(chan Contact)
	defer close(sendCh)


	id := NewKademliaID("1111111100000000000000000000000000000000")

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact1.CalcDistance(id)
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact2.CalcDistance(id)
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")
	contact3.CalcDistance(id)

	expContacts := []Contact{contact1, contact2, contact3}
	sort.Sort(ByDistance(expContacts))

	go func() {
		hasReturnd := false
		for {
			contact, more := <-sendCh
			if !more {
				return
			}

			if hasReturnd {
				receiveCh <- LookupResponse{[]Contact{}, contact, false}
			} else {
				hasReturnd = true
				receiveCh <- LookupResponse{[]Contact{contact2, contact3}, contact, true}
			}
		}
	}()

	contacts, contact, _ := RunLookup(id, contact1, []Contact{contact3}, sendCh, receiveCh)

	if !contact.ID.Equals(contact2.ID) {
		t.Error("Expected contact ", contact2, " got ", contact)
	}

	checkContacts(t, contacts, expContacts)
}


func TestRunLookup(t *testing.T) {
	receiveCh := make(chan LookupResponse)
	defer close(receiveCh)

	sendCh := make(chan Contact)
	defer close(sendCh)

	id := NewKademliaID("1111111100000000000000000000000000000000")

	contact1 := NewContact(NewKademliaID("2111111400000000000000000000000000000000"), "localhost:8002")
	contact1.CalcDistance(id)
	contact2 := NewContact(NewKademliaID("1111111400000000000000000000000000000000"), "localhost:8002")
	contact2.CalcDistance(id)
	contact3 := NewContact(NewKademliaID("1111111100000000000000000000000000000000"), "localhost:8002")
	contact3.CalcDistance(id)

	contact4 := NewContact(NewKademliaID("1111111500000000000000000000000000000000"), "localhost:8002")
	contact4.CalcDistance(id)

	contact5 := NewContact(NewKademliaID("1111111600000000000000000000000000000000"), "localhost:8002")
	contact5.CalcDistance(id)

	contact6 := NewContact(NewKademliaID("1111111700000000000000000000000000000000"), "localhost:8002")
	contact6.CalcDistance(id)

	expContacts := []Contact{contact1, contact2, contact3, contact4, contact5, contact6}
	sort.Sort(ByDistance(expContacts))

	go func() {
		hasReturnd := false
		for {
			contact, more := <-sendCh
			if !more {
				return
			}
			go func(contact Contact){
				if hasReturnd {
					receiveCh <- LookupResponse{[]Contact{contact1}, contact, false}
				} else {
					hasReturnd = true
					receiveCh <- LookupResponse{[]Contact{contact2, contact3, contact4, contact5, contact6}, contact, true}
				}
			}(contact)
		}
	}()

	contacts, contact, _ := RunLookup(id, contact1, []Contact{contact3}, sendCh, receiveCh)

	if !contact.ID.Equals(contact2.ID) {
		t.Error("Expected contact ", contact2, " got ", contact)
	}

	checkContacts(t, contacts, expContacts[:k])
}

func checkContacts(t *testing.T, contacts []Contact, expContacts []Contact) {
	if len(contacts) != len(expContacts) {
		t.Error("Expected ",  len(expContacts), " contacts got ", len(contacts))
	}

	for i := 0; i < len(expContacts); i++ {
		if !contacts[i].ID.Equals(expContacts[i].ID) {
			t.Error("Expected contact ", expContacts[i], " got ", contacts[i])
		}
	}
}

