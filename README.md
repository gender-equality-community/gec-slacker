# GEC Slacker

Take messages from an XSTREAM, and post to slack.

For each unique recipient ID, create a new slack channel (if one doesn't exist), post the message to
that channel (with a string at the end containing the derived sentiment).

For new recipients, update a central announce channel. For updated messages, simply update a logs
channel.
