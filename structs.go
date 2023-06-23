package main

type ChangeFriends struct {
	Persons []Person
	Changes []Change
}

type ResponseFriends struct {
	Friends Friends `json:"response"`
}

type Friends struct {
	Count 	int 		`json:"count"`
	Persons []Person 	`json:"items"`
}

type Person struct {
	Id 			uint32 	`json:"id"`
	FirstName 	string 	`json:"first_name"`
	LastName 	string 	`json:"last_name"`
}

type Change struct {
	Person 	Person
	Data 	string
	Status 	string
}