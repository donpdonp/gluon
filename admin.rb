puts Neur0n::machine_list.inspect

module Neur0n
  def self.dispatch(msg)
    puts "admin.rb dispatch #{msg.inspect}"
    if msg['type'] == 'vm.add'
      if msg['name']
        puts "Adding machine #{msg['name']}"
        idx = Neur0n::machine_add(msg['name'])
        if idx && msg["url"]
          puts "loading #{msg['url']}"
          code = Neur0n::http_get(msg['url'])
          Neur0n::machine_eval(msg['name'], code)
        end
      end
    end
  end
end
