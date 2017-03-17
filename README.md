# troll_captcha

## Objective Complete
access troll captcha at http://ek2y.com
Reload browser to see all random troll captchas

        # You can run server on Mac OSX by running main, (you will need to have go installed on your local environment)
        # Run/Start the server using go run which runs on localhost:8000/
        # By default this will use local files to create captcha text
        
        go run main.go

        # You can add the following flag -captcha_numb to determine the captcha text size
        # If this flag is set and greater than 0, it will fetch for that many Chuck Norris
        # jokes as the captcha text
        
        go run main.go -captcha_numb=100

        # OR if you don't have go installed you can just run the binary by running the included
        #main executable, the version included on github is compiled for Linux distros
        #You will also need to run this exectutable from the same directory as the text and 
        #template directories
        
        ./main

        # Run the test suite from the models directory
       
        go test

## Assumptions/Decisions made

### 1. Respond to a orginal client HTTP request

1. The exclusion list sent back as response should only include words within the original text
2. Troll Captchas with only one unique word should have an empty exclusion list
3. Exclusion lists can't be as large as the texts in the captcha, at max one less than total words.  This prevents the troll from submitting all ignored fields; therefore, not requiring any counting, just submitting blank forms.
4. Captchas with more than one word will have exclusion list form 1 to 5 at random
5. Each element within exclusion list is random, but only random when created and cached (not per request), meaning each Captcha will have the same exlusion list until server restarted.  Depending on future, client demands, this might change so that each request triggers a random exclusion list.  Made this original decision based on performance and security concerns.  The assumption may be that if a troll cracks a Captcha, he will not share with his friends.
6. Only words starting with a letter will be included in the exclusion list
7. No duplicates in the exclusion list
8. Punctuation and special characters are striped from both the front and end of words
9. Special characters are allowed in the middle of a word (for example "real-time") but space will separate the text

### 1. Receive a client HTTP request with counts
1. Exclusion list can be in any order
2. Words to be excluded/ignored must either be submitted with no value or a value of 0.  With more time, the ui should have a way to remove input for excluded words as determined by troll
3. Only implemented a web client not a json api, with more time could create api with implemented endpoints for passing following params: original text, included words, excluded words, would be similar to implement but the object mapping would require some changes
4. Server responds with success if captcha is correct and failure if not, both views complete with a link to try again
5. Every word must contain a count, just submitting a sum of all the counts is not secure enough
6. Sending duplicate included or or excluded words results in failure
7. Client must send original text in original form or will fail
8. Verification is based on the preloaded local texts and what is submitted by troll.  It is stateless, as the server uses the text submitted unique id (a hashed id of original text) and compares that unique text result on what the client should have submitted.  The server does not depend on the state of any previous requests.  That being said, because of the randomness in creating the captcha, once the server restarts new captcha are created and therefore can't compare older captcha.


## Documentation

All code is documented within source code, as per common golang.

## Tests
Only provided unit testing on the troll_captcha model.  With more, time would add some integration/functional testing to test more of the web request transactions and responses