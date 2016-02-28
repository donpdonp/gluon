
[feb 2016]
bus problem:

  * node nanomsg wont listen to mango, possibly wont listen to itself
  * nanomsg bus echo-back makes it too easy to form a msg loop in scripts
  * redis pub/sub lack of echo-back means scheme for script to emit vm.add doesnt work because vm container doesnt hear it, since container sent it

