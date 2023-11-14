function main()
	 initSystem = GetInitSystem()
	 distro = GetDistro()
	 return initSystem ~= "" and distro.Versio ~= "" and distro.Name ~= ""
end
