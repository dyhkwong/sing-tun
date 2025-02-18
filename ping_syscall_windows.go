package tun

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output ping_zsyscall_windows.go ping_syscall_windows.go

// https://learn.microsoft.com/zh-tw/windows/win32/api/icmpapi/nf-icmpapi-icmpcreatefile
//sys icmpCreateFile() (handle windows.Handle, err error) [failretval==windows.InvalidHandle] = iphlpapi.IcmpCreateFile

// https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmp6createfile
//sys icmp6CreateFile() (handle windows.Handle, err error) [failretval==windows.InvalidHandle] = iphlpapi.Icmp6CreateFile

// https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpsendecho2ex
//sys icmpSendEcho2Ex(handle windows.Handle, event windows.Handle, apcRoutine uintptr, apcContext uintptr, sourceAddress *[4]byte, destinationAddress *[4]byte, requestData []byte, requestOptions *IPOptionInformation, replyBuffer []byte, timeout uint32) (err error) = iphlpapi.IcmpSendEcho2Ex

// https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmp6sendecho2
//sys icmp6SendEcho2(handle windows.Handle, event windows.Handle, apcRoutine uintptr, apcContext uintptr, sourceAddress *windows.RawSockaddrInet6, destinationAddress *windows.RawSockaddrInet6, requestData []byte, requestOptions *IPOptionInformation, replyBuffer []byte, timeout uint32) (err error) = iphlpapi.Icmp6SendEcho2

// https://learn.microsoft.com/en-us/windows/win32/api/icmpapi/nf-icmpapi-icmpparsereplies
//sys icmpParseReplies(replyBuffer []byte) (replies uint32, err error) [failretval == 0] = iphlpapi.IcmpParseReplies

// https://learn.microsoft.com/zh-tw/windows/win32/api/icmpapi/nf-icmpapi-icmpclosehandle
//sys icmpCloseHandle(icmpHandle windows.Handle) (err error) [failretval == 0] = iphlpapi.IcmpCloseHandle
