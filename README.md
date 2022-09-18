[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=gender-equality-community_gec-slacker&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=gender-equality-community_gec-slacker)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=gender-equality-community_gec-slacker&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=gender-equality-community_gec-slacker)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=gender-equality-community_gec-slacker&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=gender-equality-community_gec-slacker)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=gender-equality-community_gec-slacker&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=gender-equality-community_gec-slacker)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=gender-equality-community_gec-slacker&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=gender-equality-community_gec-slacker)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=gender-equality-community_gec-slacker&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=gender-equality-community_gec-slacker)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=gender-equality-community_gec-slacker&metric=bugs)](https://sonarcloud.io/summary/new_code?id=gender-equality-community_gec-slacker)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker?ref=badge_shield)

---
# GEC Slack Bot

The GEC Slacker project does two things:

1. It receives labelled messages from a redis XStream (see: [here](https://github.com/gender-equality-community/gec-bot#gender-equality-community-whatsapp-bot) for context and a diagram) and sends them to slack
2. It receives responses from slack and sends them back to users, via redis, anonymously back to users too

Each individual respondant, as determined by the `Message.ID` field set in the [gec-bot](https://github.com/gender-equality-community/gec-bot), gets their own slack channel, allowing us to do clever things like track user conversations and responses.

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fgender-equality-community%2Fgec-slacker?ref=badge_large)
