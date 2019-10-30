package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nlopes/slack"
	"github.com/nlopes/slack/slackevents"
)

// You more than likely want your "Bot User OAuth Access Token" which starts with "xoxb-"
var token = os.Getenv("token")
var api = slack.New(token,
	slack.OptionDebug(true),
	slack.OptionLog(
		log.New(os.Stdout, "slack-bot: ",
			log.Lshortfile|log.LstdFlags)),)
//"xoxb-2152601087-518569019028-puIGyAd0NLaxFETP3Y3DtMDO" --test

func actionHandler(w http.ResponseWriter, r *http.Request){

	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(r.FormValue("payload")), &payload)
	if err != nil {
		fmt.Printf("Could not parse action response JSON: %v", err)
	}
	log.Printf("My payload %+v\n", payload)

}

func handler(w http.ResponseWriter, r *http.Request) {
	var request []string
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	log.Printf("My request headers %v\n", request)

	defer r.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	log.Printf("My request body %s\n", body)
	secretsVerifier, err := slack.NewSecretsVerifier(r.Header, "3965e3d97b691ed7f4e254b9735d23b8")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	secretsVerifier.Write([]byte(body))
	if err := secretsVerifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
	}

	log.Printf("My parsed Event %+v\n", eventsAPIEvent)


	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	}
	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			appMentionEvent := innerEvent.Data.(*slackevents.AppMentionEvent)
			channel, ts, err := api.PostMessage(ev.Channel, slack.MsgOptionBlocks(actionInteractionStart() ...),
				slack.MsgOptionTS(appMentionEvent.TimeStamp))
			log.Printf("Successful post to channel %v at %v\n", channel, ts)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func actionInteractionStart() []slack.Block  {
	var blocks []slack.Block

	headerText := slack.NewTextBlockObject("mrkdwn", "Testing with interactive actions", false, false)

	// Create the header section
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// Shared Divider
	divSection := slack.NewDividerBlock()

	chooseBtnText := slack.NewTextBlockObject("plain_text", "You can add a button alongside text in your message.", true, false)
	chooseBtnEle := slack.NewButtonBlockElement("", "yes", chooseBtnText)

	optionOneText := slack.NewTextBlockObject("mrkdwn", "Test", false, false)
	optionOneSection := slack.NewSectionBlock(optionOneText, nil, slack.NewAccessory(chooseBtnEle))

	blocks = append(blocks,
		headerSection,
		divSection,
		optionOneSection,
		)


	return blocks
}

func exampleEasy() []slack.Block{
	var blocks []slack.Block

	headerText := slack.NewTextBlockObject("mrkdwn", "We found *100 Clusters* for profile *dev*", false, false)

	// Create the header section
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// Shared Divider
	divSection := slack.NewDividerBlock()

	clusterOne := slack.NewTextBlockObject("mrkdwn", "*scus-dev-a2*", false, false)

	clusterOneProfile := slack.NewTextBlockObject("plain_text", "Profile: dev", true, false)
	clusterOneSite := slack.NewTextBlockObject("plain_text", "Site: Azure", true, false)

	clusterOneSection := slack.NewSectionBlock(clusterOne, nil, nil)
	clusterOneContext := slack.NewContextBlock("", []slack.MixedElement{clusterOneProfile, clusterOneSite}...)


	clusterTwo := slack.NewTextBlockObject("mrkdwn", "*scus-dev-a3*", false, false)

	clusterTwoProfile := slack.NewTextBlockObject("plain_text", "Profile: dev", true, false)
	clusterTwoSite := slack.NewTextBlockObject("plain_text", "Site: Azure", true, false)

	clusterTwoSection := slack.NewSectionBlock(clusterTwo, nil, nil)
	clusterTwoContext := slack.NewContextBlock("", []slack.MixedElement{clusterTwoProfile, clusterTwoSite}...)

	blocks = append(blocks,
		headerSection,
		divSection,
		clusterOneSection,
		clusterOneContext,
		divSection,
		clusterTwoSection,
		clusterTwoContext,
		)

	//return b
	return blocks

}



func exampleFive() []slack.Block{

	var blocks []slack.Block
	// Build Header Section Block, includes text and overflow menu

	headerText := slack.NewTextBlockObject("mrkdwn", "We found *205 Hotels* in New Orleans, LA from *12/14 to 12/17*", false, false)

	// Build Text Objects associated with each option
	overflowOptionTextOne := slack.NewTextBlockObject("plain_text", "Option One", false, false)
	overflowOptionTextTwo := slack.NewTextBlockObject("plain_text", "Option Two", false, false)
	overflowOptionTextThree := slack.NewTextBlockObject("plain_text", "Option Three", false, false)

	// Build each option, providing a value for the option
	overflowOptionOne := slack.NewOptionBlockObject("value-0", overflowOptionTextOne)
	overflowOptionTwo := slack.NewOptionBlockObject("value-1", overflowOptionTextTwo)
	overflowOptionThree := slack.NewOptionBlockObject("value-2", overflowOptionTextThree)

	// Build overflow section
	overflow := slack.NewOverflowBlockElement("", overflowOptionOne, overflowOptionTwo, overflowOptionThree)

	// Create the header section
	headerSection := slack.NewSectionBlock(headerText, nil, slack.NewAccessory(overflow))

	// Shared Divider
	divSection := slack.NewDividerBlock()

	// Shared Objects
	locationPinImage := slack.NewImageBlockElement("https://api.slack.com/img/blocks/bkb_template_images/tripAgentLocationMarker.png", "Location Pin Icon")

	// First Hotel Listing
	hotelOneInfo := slack.NewTextBlockObject("mrkdwn", "*<fakeLink.toHotelPage.com|Windsor Court Hotel>*\n★★★★★\n$340 per night\nRated: 9.4 - Excellent", false, false)
	hotelOneImage := slack.NewImageBlockElement("https://api.slack.com/img/blocks/bkb_template_images/tripAgent_1.png", "Windsor Court Hotel thumbnail")
	hotelOneLoc := slack.NewTextBlockObject("plain_text", "Location: Central Business District", true, false)

	hotelOneSection := slack.NewSectionBlock(hotelOneInfo, nil, slack.NewAccessory(hotelOneImage))
	hotelOneContext := slack.NewContextBlock("", []slack.MixedElement{locationPinImage, hotelOneLoc}...)

	// Second Hotel Listing
	hotelTwoInfo := slack.NewTextBlockObject("mrkdwn", "*<fakeLink.toHotelPage.com|The Ritz-Carlton New Orleans>*\n★★★★★\n$340 per night\nRated: 9.1 - Excellent", false, false)
	hotelTwoImage := slack.NewImageBlockElement("https://api.slack.com/img/blocks/bkb_template_images/tripAgent_2.png", "Ritz-Carlton New Orleans thumbnail")
	hotelTwoLoc := slack.NewTextBlockObject("plain_text", "Location: French Quarter", true, false)

	hotelTwoSection := slack.NewSectionBlock(hotelTwoInfo, nil, slack.NewAccessory(hotelTwoImage))
	hotelTwoContext := slack.NewContextBlock("", []slack.MixedElement{locationPinImage, hotelTwoLoc}...)

	// Third Hotel Listing
	hotelThreeInfo := slack.NewTextBlockObject("mrkdwn", "*<fakeLink.toHotelPage.com|Omni Royal Orleans Hotel>*\n★★★★★\n$419 per night\nRated: 8.8 - Excellent", false, false)
	hotelThreeImage := slack.NewImageBlockElement("https://api.slack.com/img/blocks/bkb_template_images/tripAgent_3.png", "https://api.slack.com/img/blocks/bkb_template_images/tripAgent_3.png")
	hotelThreeLoc := slack.NewTextBlockObject("plain_text", "Location: French Quarter", true, false)

	hotelThreeSection := slack.NewSectionBlock(hotelThreeInfo, nil, slack.NewAccessory(hotelThreeImage))
	hotelThreeContext := slack.NewContextBlock("", []slack.MixedElement{locationPinImage, hotelThreeLoc}...)

	// Action button
	btnTxt := slack.NewTextBlockObject("plain_text", "Next 2 Results", false, false)
	nextBtn := slack.NewButtonBlockElement("", "click_me_123", btnTxt)
	actionBlock := slack.NewActionBlock("", nextBtn)

	blocks = append(blocks, 
	headerSection,
		divSection,
		hotelOneSection,
		hotelOneContext,
		divSection,
		hotelTwoSection,
		hotelTwoContext,
		divSection,
		hotelThreeSection,
		hotelThreeContext,
		divSection,
		actionBlock)

	// Build Message with blocks created above
   msg := slack.NewBlockMessage(
		headerSection,
		divSection,
		hotelOneSection,
		hotelOneContext,
		divSection,
		hotelTwoSection,
		hotelTwoContext,
		divSection,
		hotelThreeSection,
		hotelThreeContext,
		divSection,
		actionBlock,
	)

	b, err := json.MarshalIndent(msg, "", "     ")
	if err != nil {
		log.Print(err)
		
	}

	log.Print(string(b))

return blocks


}

func main() {
	http.HandleFunc("/events-endpoint", handler)
	http.HandleFunc("/actions", actionHandler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
