package mgostore

import (
	"github.com/RangelReale/osin"

	"encoding/json"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// collection names for the entities
const (
	CLIENT_COL    = "clients"
	AUTHORIZE_COL = "authorizations"
	ACCESS_COL    = "accesses"
)

const REFRESHTOKEN = "refreshtoken"

type MongoStorage struct {
	dbName  string
	session *mgo.Session
}

func New(session *mgo.Session, dbName string) *MongoStorage {
	storage := &MongoStorage{dbName, session}
	index := mgo.Index{
		Key:        []string{REFRESHTOKEN},
		Unique:     false, // refreshtoken is sometimes empty
		DropDups:   false,
		Background: true,
		Sparse:     true,
	}
	accesses := storage.session.DB(dbName).C(ACCESS_COL)
	err := accesses.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
	return storage
}

func (store *MongoStorage) Clone() osin.Storage {
	return store
}

func (store *MongoStorage) Close() {

}

func (store *MongoStorage) GetClient(id string) (osin.Client, error) {
	session := store.session.Copy()
	defer session.Close()
	clients := session.DB(store.dbName).C(CLIENT_COL)
	client := new(osin.DefaultClient)
	err := clients.FindId(id).One(client)
	return client, err
}

func (store *MongoStorage) SetClient(id string, client osin.Client) error {
	session := store.session.Copy()
	defer session.Close()
	clients := session.DB(store.dbName).C(CLIENT_COL)
	_, err := clients.UpsertId(id, client)
	return err
}

func (store *MongoStorage) SaveAuthorize(data *osin.AuthorizeData) error {
	session := store.session.Copy()
	defer session.Close()
	authorizations := session.DB(store.dbName).C(AUTHORIZE_COL)
	_, err := authorizations.UpsertId(data.Code, data)
	return err
}

func (store *MongoStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	session := store.session.Copy()
	defer session.Close()
	authorizations := session.DB(store.dbName).C(AUTHORIZE_COL)
	authData := osin.AuthorizeData{
		Client: new(osin.DefaultClient),
	}
	generic := make(bson.M)
	if err := authorizations.FindId(code).One(&generic); err != nil {
		return &authData, err
	}

	j1, err := json.Marshal(generic)
	if err != nil {
		return &authData, err
	}

	err = json.Unmarshal(j1, &authData)

	return &authData, err
}

func (store *MongoStorage) RemoveAuthorize(code string) error {
	session := store.session.Copy()
	defer session.Close()
	authorizations := session.DB(store.dbName).C(AUTHORIZE_COL)
	return authorizations.RemoveId(code)
}

func (store *MongoStorage) SaveAccess(data *osin.AccessData) error {
	session := store.session.Copy()
	defer session.Close()
	accesses := session.DB(store.dbName).C(ACCESS_COL)
	_, err := accesses.UpsertId(data.AccessToken, data)
	return err
}

func (store *MongoStorage) LoadAccess(token string) (*osin.AccessData, error) {
	session := store.session.Copy()
	defer session.Close()
	accesses := session.DB(store.dbName).C(ACCESS_COL)
	accData := new(osin.AccessData)
	err := accesses.FindId(token).One(accData)
	return accData, err
}

func (store *MongoStorage) RemoveAccess(token string) error {
	session := store.session.Copy()
	defer session.Close()
	accesses := session.DB(store.dbName).C(ACCESS_COL)
	return accesses.RemoveId(token)
}

func (store *MongoStorage) LoadRefresh(token string) (*osin.AccessData, error) {
	session := store.session.Copy()
	defer session.Close()
	accesses := session.DB(store.dbName).C(ACCESS_COL)
	accData := osin.AccessData{Client: &osin.DefaultClient{}}
	genericAccessData := make(bson.M)
	if err := accesses.Find(bson.M{REFRESHTOKEN: token}).One(&genericAccessData); err != nil {
		return &accData, err
	}

	j1, err := json.Marshal(genericAccessData)
	if err != nil {
		return &accData, err
	}
	err = json.Unmarshal(j1, &accData)
	if err != nil {
		return &accData, err
	}

	return &accData, err
}
func (store *MongoStorage) RemoveRefresh(token string) error {
	session := store.session.Copy()
	defer session.Close()
	accesses := session.DB(store.dbName).C(ACCESS_COL)
	return accesses.Update(bson.M{REFRESHTOKEN: token}, bson.M{
		"$unset": bson.M{
			REFRESHTOKEN: 1,
		}})
}
