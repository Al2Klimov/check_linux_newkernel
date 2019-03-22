#!/usr/bin/perl

{
	local $/ = undef;
	local @ARGV = ('/proc/uptime');
	$_ = <>
}

if (/^(\d+)/) {
	my $bootTime = time() - $1;
	my $mtime = $bootTime + 30;

	sleep 60;
	open(my $f, '>', '/boot/vmlinuz');
	utime($mtime, $mtime, '/boot/vmlinuz');

	for (;;) {
		sleep 60;
		$mtime = $bootTime - 30;
		utime($mtime, $mtime, '/boot/vmlinuz');

		sleep 60;
		$mtime = $bootTime + 30;
		utime($mtime, $mtime, '/boot/vmlinuz')
	}
}
