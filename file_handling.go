package main

import (
	"encoding/json"
	"os"
	"time"
	"vk_friends/logger"
)

func startFriendList(r *ResponseFriends) ChangeFriends {
	filename := "data.txt"
	data := ChangeFriends{}

	byte, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		fileCreate(filename)
		byte, err = os.ReadFile(filename)
		if err != nil {
			logger.Error.Fatalln(err)
		}
	} else if err != nil {
		logger.Error.Fatalln(err)
	}

	if len(byte) != 0 {
		jErr := json.Unmarshal(byte, &data)
		if jErr != nil {
			logger.Error.Fatalln("Can't read correct data from file", filename)
		}
		// Сравнение данных
		compareData(r, &data)
	} else {
		// Если файл с данными пустой, записывается текущий список друзей
		data.Persons = r.Friends.Persons
	}
	WriteFile(filename, &data)
	return data
}

func compareData(r *ResponseFriends, f *ChangeFriends) {
	// Карта с id текущих друзей
	respData := make(map[uint32]Change)
	for _, val := range r.Friends.Persons {
		respData[val.Id] = Change{val, "", "added"}
	}
	// Сравнение старого списка друзей с новым
	for i := range f.Persons {
		key := f.Persons[i].Id
		_ , found := respData[key]
		if found {
			delete(respData, key)
		} else {
			respData[key] = Change{f.Persons[i], "", "deleted"}
		}
	}
	// Добавление изменений
	for _, val := range respData {
		val.Data = time.Now().Format("01.02.2006")
		f.Changes = append(f.Changes, val)
	}
	// Перезаписывание списка друзей
	f.Persons = r.Friends.Persons
}

func fileCreate(filename string) {
	_, fileErr := os.Create(filename)
	if fileErr != nil {
		logger.Error.Fatalln(fileErr)
	}
	logger.Info.Println("File", filename, "created")
}

func WriteFile(filename string, r *ChangeFriends) {
	// Запись данных
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		logger.Error.Fatalln(err)
	}
	defer file.Close()
	byteData, _ := json.MarshalIndent(r, "", "")
	count, _ := file.Write(byteData)
	logger.Info.Println(count, "bytes was written to file", filename)
}

func dataAnalysis(bytes []byte) (data ChangeFriends, err error) {
	response := ResponseFriends{}
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return data, err
	}

	data = startFriendList(&response)
	return data, err
}