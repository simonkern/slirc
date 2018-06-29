package slack

import (
	"encoding/json"
	"testing"
)

func setup(t *testing.T) (sc *SlackClient) {
	sc = NewSlackClient("foobar")
	var apiResp SlackAPIResponse
	if err := json.Unmarshal(rawAPIResp, &apiResp); err != nil {
		t.Fatal(err)
	}
	sc.bookKeeping(&apiResp)
	return sc
}

/*
func TestHandleFunc(t *testing.T) {
	sc := setup(t)

	want := int64(3)

	hf := func(slc *SlackClient, e *Event) {
		slc.nextID++
	}

	sc.HandleFunc("test1", hf)
	sc.HandleFunc("test2", hf)
	sc.HandleFunc("test3", hf)

	sc.disPatchHandlers(&Event{Type: "test1"})
	sc.disPatchHandlers(&Event{Type: "test2"})
	sc.disPatchHandlers(&Event{Type: "test3"})
	time.Sleep(2)
	if sc.nextID != want {
		t.Logf("HandleFunc mechanic failed, sc.nextID value - expected: (%v), got (%v)", want, sc.nextID)
		t.Fail()
	}
}*/

func TestBookKeeping(t *testing.T) {
	sc := setup(t)

	for k, v := range wantChan {
		chan1, ok1 := sc.chanIDMap[k]
		chan2, ok2 := sc.chanMap[v]
		if !ok1 || !ok2 || chan1 != chan2 {
			t.FailNow()
		}
		if chan1.ID != k || chan1.Name != v {
			t.Fail()
		}
	}

	for k, v := range wantUser {
		user, ok := sc.userIDMap[k]
		if !ok {
			t.FailNow()
		}
		if user.ID != k || user.Profile.DisplayName != v {
			t.Fail()
		}

	}
}

var wantUser map[string]string = map[string]string{
	"U11A2B8C1": "testorizor1",
	"U11A2BBCK": "testorizor2",
	"U11A2BB4P": "testorizor3",
	"U11A2BZTT": "testorizor4",
	"U11A2B9N9": "litabot",
	"U11A2BCRJ": "testorizor5",
	"U11A2BKMR": "testorizor6",
	"U11A2BRKS": "slirctest",
	"U11A2BMFY": "testorizor7",
}

var wantChan map[string]string = map[string]string{
	"C11JBA78E": "slirctest",
	"C03JAPEHJ": "dev",
	"C0BD11R1N": "devtest",
}

var rawAPIResp []byte = []byte(`{  
   "ok":true,
   "self":{  
      "id":"U11D00T0",
      "name":"SLIRC",
      "created":1433374756,
      "manual_presence":"active"
   },
   "channels":[  
      {  
         "id":"C11JBA78E",
         "name":"slirctest",
         "is_channel":true,
         "created":1423105695,
         "creator":"U03J7E8C1",
         "is_archived":false,
         "is_general":false,
         "has_pins":false,
         "is_member":true,
         "last_read":"1433375718.000679",
         "unread_count":1000,
         "unread_count_display":1000,
         "members":[  
            "U11A2B8C1",
            "U11A2BBCK",
            "U11A2BB4P",
            "U11A2BZTT",
            "U11A2B9N9",
            "U11A2BCRJ",
            "U11A2BKMR",
            "U11A2BRKS",
            "U11A2BMFY"
         ]
      },
      {  
         "id":"C03JAPEHJ",
         "name":"dev",
         "is_channel":true,
         "created":1423101439,
         "creator":"U03J7E8C1",
         "is_archived":true,
         "is_general":false,
         "has_pins":false,
         "is_member":false
      },
      {  
         "id":"C0BD11R1N",
         "name":"devtest",
         "is_channel":true,
         "created":1443387822,
         "creator":"U03JDH2EP",
         "is_archived":false,
         "is_general":false,
         "has_pins":false,
         "is_member":false
      }
   ],
   "cache_ts":1443527725,
   "users":[  
      {  
         "id":"U11A2B8C1",
         "name":"tester1",
         "profile": {
          "display_name": "testorizor1"
         },
         "deleted":false,
         "status":null,
         "color":"3c989f",
         "real_name":"Christoph Rackwitz",
         "tz":"Europe\/Amsterdam",
         "tz_label":"Central European Summer Time",
         "tz_offset":7200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "has_files":false,
         "presence":"away"
      },
     
      {  
         "id":"U11A2BBCK",
         "name":"tester2",
         "profile": {
          "display_name": "testorizor2"
         },
         "deleted":false,
         "status":null,
         "color":"9d8eee",
         "real_name":"Joshua Fuerste",
         "tz":"Europe\/Amsterdam",
         "tz_label":"Central European Summer Time",
         "tz_offset":7200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "has_files":true,
         "presence":"away"
      },
 
      {  
         "id":"U11A2BB4P",
         "name":"tester3",
         "profile": {
          "display_name": "testorizor3"
         },
         "deleted":false,
         "status":null,
         "color":"9f69e7",
         "tz":"Europe\/Amsterdam",
         "tz_label":"Central European Summer Time",
         "tz_offset":7200,
         "is_admin":true,
         "is_owner":true,
         "is_primary_owner":true,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "has_files":true,
         "presence":"active"
      },
      {  
         "id":"U11A2BZTT",
         "name":"tester4",
         "profile": {
          "display_name": "testorizor4"
         },
         "deleted":false,
         "status":null,
         "color":"d1707d",
         "tz":"Europe\/Amsterdam",
         "tz_label":"Central European Summer Time",
         "tz_offset":7200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "has_files":false,
         "presence":"away"
      },
      {  
         "id":"U11A2B9N9",
         "name":"litabot",
         "profile": {
          "display_name": "litabot"
         },
         "deleted":false,
         "status":null,
         "color":"7d414c",
         "tz":null,
         "tz_label":"Pacific Daylight Time",
         "tz_offset":-25200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":true,
         "has_files":false,
         "presence":"active"
      },
      {  
         "id":"U11A2BCRJ",
         "name":"tester5",
         "profile": {
          "display_name": "testorizor5"
         },
         "deleted":false,
         "status":null,
         "color":"5b89d5",
         "tz":"Europe\/Amsterdam",
         "tz_label":"Central European Summer Time",
         "tz_offset":7200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "has_files":true,
         "presence":"away"
      },

      {  
         "id":"U11A2BKMR",
         "name":"tester6",
         "profile": {
          "display_name": "testorizor6"
         },
         "deleted":false,
         "status":null,
         "color":"de5f24",
         "tz":"Europe\/Amsterdam",
         "tz_label":"Central European Summer Time",
         "tz_offset":7200,
         "is_admin":true,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "has_files":true,
         "presence":"active"
      },
      {  
         "id":"U11A2BRKS",
         "name":"slirctest",
         "profile": {
          "display_name": "slirctest"
         },
         "deleted":false,
         "status":null,
         "color":"7d414c",
         "real_name":"",
         "tz":null,
         "tz_label":"Pacific Daylight Time",
         "tz_offset":-25200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":true,
         "has_files":false,
         "presence":"away"
      },

      {  
         "id":"U11A2BMFY",
         "name":"tester7",
         "profile": {
          "display_name": "testorizor7"
         },
         "deleted":false,
         "status":null,
         "color":"a2a5dc",
         "tz":"Europe\/Amsterdam",
         "tz_label":"Central European Summer Time",
         "tz_offset":7200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "has_files":true,
         "presence":"away"
      },
      {  
         "id":"USLACKBOT",
         "name":"slackbot",
         "profile": {
          "display_name": "slackbot"
         },
         "deleted":false,
         "status":null,
         "color":"757575",
         "tz":null,
         "tz_label":"Pacific Daylight Time",
         "tz_offset":-25200,
         "is_admin":false,
         "is_owner":false,
         "is_primary_owner":false,
         "is_restricted":false,
         "is_ultra_restricted":false,
         "is_bot":false,
         "presence":"active"
      }
   ],
   "cache_version":"v11-mouse",
   "cache_ts_version":"v1-cat",
   "bots":[  
      {  
         "id":"B088HCDS8",
         "deleted":false,
         "name":"snarkov",
         "icons":{  
            "image_48":"https:\/\/slack.global.ssl.fastly.net\/4324\/img\/services\/incoming-webhook_48.png"
         }
      },
      {  
         "id":"B0BD6AXDY",
         "deleted":false,
         "name":"bot",
         "icons":{  
            "image_48":"https:\/\/slack.global.ssl.fastly.net\/93ed\/img\/services\/bots_48.png"
         }
      },
      {  
         "id":"B03J9793P",
         "deleted":true,
         "name":"IRCBot",
         "icons":{  
            "image_48":"https:\/\/slack.global.ssl.fastly.net\/4324\/img\/services\/incoming-webhook_48.png"
         }
      },
      {  
         "id":"B03J8N95H",
         "deleted":true,
         "name":"incoming-webhook",
         "icons":{  
            "image_48":"https:\/\/slack.global.ssl.fastly.net\/4324\/img\/services\/incoming-webhook_48.png"
         }
      }
   ],
   "url":"wss:\/\/ws314.slack-msgs.com\/websocket\/a-AAAaAAAaAAAAaAA1AAaaaAA1_-x1CuO7QZjt_cf8pP7GMfUsjODU6KKH8DBwSui_azNEMcFwYAAAa_AAaaa11aAAAaAaaAA1aAa1aAaa0="
}`)
