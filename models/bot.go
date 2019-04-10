package models

import "errors"

type IsLeisure bool
type BotName string

type Bot map[BotName]IsLeisure

type BotManager struct{
  BotMap Bot
}

func (bm *BotManager) SetBots(botNames []string){
  bm.BotMap = make(map[BotName]IsLeisure)
  for _,name := range botNames{
    bm.BotMap[BotName(name)] = true
  }
}

func (bm *BotManager) GetLeisureBot() (botName string, err error){
  for name,isLeisure := range bm.BotMap{
    if isLeisure{
      bm.BotMap[name] = false
      return string(name), nil
    }
  }
  return "", errors.New("There is no leisure bot")
}

func (bm *BotManager) ReleaseBot(name string){
  bm.BotMap[BotName(name)] = IsLeisure(true)
}

