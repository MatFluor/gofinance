# GoFinance

a simple web application to aid in financial control through a daily "magic number"

## Installation

1. make sure your $GOPATH is set
2. go get github.com/julienschmidt/httprouter
3. go build

## Usage

1. Enter your fixed expenses (like rent or other stuff that is not mutable and recurrent)
2. This calculates your "magic number", your daily amount of money you can spend
3. Enter each new expense you have. There is no categorization, since I always found those to be too tedious to make it a habit
4. Have control over your finances - purchase for purchase, day after day!

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D

## History

* Late 2014 - First implementation in MeteorJS
* October 2016 - complete Rewrite in GO for fun and learning

## Credits

Thanks to Alex Recker who wrote his blogpost "Our New Sid Meier's Civilization Inspired Budget".
If you haven't read it, do it now!
Also thanks to Julien Schmidt for his httprouter package.

## License

GoFinance is licensed under the GPLv3, see LICENSE.md
