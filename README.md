# VaxTracker - A tool to track vaccine availability

VaxTracker is an extensible tool to track the availability of vaccine (covid to begin with) across different sources (cowin to start with) and to be notified over different tools like email or slack etc. 

### What is included in this initial release : -
* Track vaccine availability from CoWin for districts that we are interested in.
* Notification is sent to "formspree" as a webhook which in return sends mails to subscribed targets. Can even send notifications to Slack or Trello through formspree.
* Checks availability based on 5 star Unix standard cron notation.

### What makes this special?
* Pretty extensible as in can add different sources to check for availability. Easily extendable by implementing simple methods. Has been developed as a platform so that it can be extended for different use cases around vaccinations for current times as well for the future.
* Similarly can easily extend the different channels through which a user can be notified. This too can be extended by implementing simple methods.


### How to build it?
* Execute the following command once you have Go installed in your system.

```
GOOS=linux go build -o vaxTracker
```

### How to run it?
* It requires a config file at the root folder where the binary rests with the file name of "config.env".
  A sample config.env file would contain the following data to run the cron once everyday at 12 am.

```
NOTIFIER__FORMSPREE__FORM_ID=xgeraapg
SCHEDULER__DISTRICTS=286 289
DEFAULT_CRON="0 0 * * *"
```

#### Future Improvements

* Add support to include other notifiers like Slack or maybe SMS.
* Add additional APIs to do on demand querying.