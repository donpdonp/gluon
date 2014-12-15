puts "I am admin.rb"

module Neur0n
  def self.dispatch(msg)
    puts "I AM DIspatch #{msg.inspect}"
    if msg['name']
      puts "Adding machine #{msg['name']}"
      Neur0n::machine_add(msg['name'])
      if msg["url"]
        puts "loading #{msg['url']}"
        code = Neur0n::http_get(msg['url'])
        puts "Admin got url body of #{code}"
      end
    end
  end
end
