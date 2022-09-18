# GEC Slacker
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker?ref=badge_shield)


Take messages from an XSTREAM, and post to slack.

For each unique recipient ID, create a new slack channel (if one doesn't exist), post the message to
that channel (with a string at the end containing the derived sentiment).

For new recipients, update a central announce channel. For updated messages, simply update a logs
channel.


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker?ref=badge_large)