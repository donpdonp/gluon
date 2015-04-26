class MasterControlProgram
  def initialize
    @machines = {}
  end

  def dispatch(msg)
    puts "admin.rb dispatch #{msg.inspect}"
    if msg['type'] == 'vm.add'
      if msg['name']
        machine = { id: newId, name: msg['name'] }
        puts "Adding machine #{machine}"
        idx = Neur0n::machine_add(machine[:id])
        @machines[machine[:id]] = machine
        if idx && msg["url"]
          machine[:url] = msg['url']
          puts "loading #{msg['url']}"
          code = Neur0n::http_get(msg['url'])
          Neur0n::machine_eval(machine[:id], code)
        end
      end
    end
    if msg['type'] == 'vm.list'
      #{machines: Neur0n::machine_list}
      puts "list #{@machines.inspect}"
      @machines
    end
  end

  def newId
    alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
    name = ""
    16.times{ name += alphabet[rand(36)]}
    name
  end
end

MCP = MasterControlProgram.new

module Neur0n
  def self.dispatch(msg)
    MCP.dispatch(msg)
  end
end
