module Neur0n
  def self.dispatch(msg)
    puts "admin.rb dispatch #{msg.inspect}"
    if msg['name']
      idx = Neur0n::machine_add(msg['name'])
      puts "Adding machine #{msg['name']} ##{idx}"
      if msg["url"]
        puts "loading #{msg['url']}"
        code = Neur0n::http_get(msg['url'])
        puts "Admin machine_eval(#{msg['name'].class}, #{code})"
        Neur0n::machine_eval(msg['name'], code)
        puts "Admin machine_eval done"
      end
    end
  end
end
