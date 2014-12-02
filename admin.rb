puts "I am admin.rb"

module Neur0n
  def self.dispatch(msg)
    puts "I AM DIspatch #{msg.inspect}"
    if msg["code"]
      puts "Adding machine mac-a"
      Neur0n::machine_add("mac-a")
      Neur0n::machine_eval("mac-a")
    end
  end
end
