package engine

//
// 主要用于无法通过监听pprof端口来查看服务器状态的场合（比如tx代理，服务器隔离程度很高），
// 例如，项目由其他公司代理运营，研发商没有权限接触到服务器，也不能在对应的服务器上开启pprof监听端口，
// 只能通过代理商提供的工具拉取日志文件，所以需要调整通常使用的pprof调试途径，改成所有调试信息都写入文件
// 然后再拉取profile文件到本地，再做分析
//

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

// allocs: A sampling of all past memory allocations
// block: Stack traces that led to blocking on synchronization primitives
// cmdline: The command line invocation of the current program
// goroutine: Stack traces of all current goroutines
// heap: A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.
// mutex: Stack traces of holders of contended mutexes
// profile: CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.
// threadcreate: Stack traces that led to the creation of new OS threads
// trace: A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.

const ( // 为了兼容windows文件命名格式，不能使用冒号
	TimeFormatLayout = "2006-01-02T15-04-05"
)

const ( // profiles types
	ProfileTypeGoroutine    = "goroutine"
	ProfileTypeThreadCreate = "threadcreate"
	ProfileTypeHeap         = "heap"
	ProfileTypeAlloc        = "allocs"
	ProfileTypeBlock        = "block"
	ProfileTypeMutex        = "mutex"
	ProfileTypeCPUProfile   = "cpuprofile" // 需要特做，不能通过pprof.Lookup直接取到
	ProfileTypeTrace        = "trace"      // 需要特做，不能通过pprof.Lookup直接取到
)

// profile类型索引profile文件后缀
var profileTypeToNameSuffix = map[string]string{
	ProfileTypeGoroutine:    ".goroutineprof",
	ProfileTypeThreadCreate: ".threadprof",
	ProfileTypeHeap:         ".heaprof",
	ProfileTypeAlloc:        ".allocprof",
	ProfileTypeBlock:        ".blockprof",
	ProfileTypeMutex:        ".mutexprof",
	ProfileTypeCPUProfile:   ".cpuprof",
	ProfileTypeTrace:        ".traceprof",
}

func profileNameTimeStrPrefix() string {
	return time.Now().Format(TimeFormatLayout)
}

// Profile 通过指定profile类型和debug值在当前目录下生成profile文件,返回文件名称
func ProfileToFile(profType string, debug int) (ret string) {
	var (
		err      error
		p        *pprof.Profile
		fileName = fileName(profType)
	)
	switch profType {
	case ProfileTypeGoroutine, ProfileTypeAlloc, ProfileTypeMutex,
		ProfileTypeHeap, ProfileTypeThreadCreate, ProfileTypeBlock: // 标准库自带函数，直接调用
		p = pprof.Lookup(profType)
	case ProfileTypeCPUProfile: // pprof.Lookup中没有现成的方法，需要自己实现
		if err = CPUProfile(30); err != nil {
			ret = err.Error()
		}
		return
	case ProfileTypeTrace: // pprof.Lookup中没有现成的方法，需要自己实现，
		if err = ProfileTrace(30); err != nil {
			ret = err.Error()
		}
		return
	default: // todo 给出错误提示
		return "type error"
	}

	if p == nil {
		return "p is nil"
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err.Error()
	}
	defer func() {
		err = f.Close()
	}()
	err = p.WriteTo(f, debug)
	if err != nil {
		return err.Error()
	}

	return fileName
}

// Profile 通过指定profile类型和debug值来搜集信息，返回[]byte格式的profile信息
func ProfileToBytes(profType string, debug int) (ret []byte, err error) {
	var (
		p      *pprof.Profile
		buffer = bytes.NewBuffer([]byte{})
	)
	switch profType {
	case ProfileTypeGoroutine, ProfileTypeAlloc, ProfileTypeMutex,
		ProfileTypeHeap, ProfileTypeThreadCreate, ProfileTypeBlock: // 标准库自带函数，直接调用
		p = pprof.Lookup(profType)
	case ProfileTypeCPUProfile: // pprof.Lookup中没有现成的方法，需要自己实现
		ret, err = CPUProfileBytes(30)
		return
	case ProfileTypeTrace: // pprof.Lookup中没有现成的方法，需要自己实现，
		ret, err = ProfileTraceToBytes(30)
		return
	default: // todo 给出错误提示
		return nil, errors.New("type error")
	}

	if p == nil {
		return nil, errors.New("p is nil")
	}

	err = p.WriteTo(buffer, debug)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// fileName 生成文件名称，格式为时间 + profType，例如：2019-04-23T16-41-55.cpuprof
func fileName(profType string) string {
	return fmt.Sprintf("%s%s", profileNameTimeStrPrefix(), profileTypeToNameSuffix[profType])
}

//---------------cpuprofile---------------

// CPUProfileBytes 生成cpu profile文件 文件名称格式为2019-04-23T16-41-55.cpuprof
func CPUProfileBytes(sec int) (ret []byte, err error) {
	buffer := bytes.NewBuffer([]byte{})
	if err = pprof.StartCPUProfile(buffer); err != nil {
		// StartCPUProfile failed, so no writes yet.
		err = errors.New(fmt.Sprintf("CPUProfile: Could not enable CPU profiling: %s", err))
		return nil, err
	}
	sleep(time.Duration(sec) * time.Second)
	pprof.StopCPUProfile()
	return buffer.Bytes(), nil
}

// CPUProfile 在当前程序运行目录下生成cpu profile文件 文件名称格式为2019-04-23T16-41-55.cpuprof
func CPUProfile(sec int) (err error) {
	infos, err := CPUProfileBytes(sec)
	if err != nil || infos == nil {
		return err
	}

	filename := fileName(ProfileTypeCPUProfile)
	f, err := os.Create(filename)
	if err != nil {
		err = errors.New(fmt.Sprintf("CPUProfile: CreateFile err %s filename %s", err, filename))
	}
	defer f.Close()
	n, err := f.Write(infos)
	if err != nil {
		err = errors.New(fmt.Sprintf("CPUProfile: WriteFile err %s filename %s num %d", err, filename, n))
	}
	return nil
}

func sleep( /*w http.ResponseWriter,*/ d time.Duration) {
	//var clientGone <-chan bool TODO 可以支持外部中断信号，例如检测到客户端断开连接，则终止信息采集
	//if cn, ok := w.(http.CloseNotifier); ok {
	//	clientGone = cn.CloseNotify()
	//}
	select {
	case <-time.After(d):
		//case <-clientGone:
	}
}

//-----------------trace-------------------
// ProfileTrace: A trace of execution of the current program. You can specify the duration in the seconds GET parameter.
// After you get the trace file, use the go tool trace command to investigate the trace.
func ProfileTrace(sec int) (err error) {
	infos, err := ProfileTraceToBytes(sec)
	filename := fileName(ProfileTypeTrace)
	f, err := os.Create(filename)
	//err = ioutil.WriteFile(filename, buffer.Bytes(), os.ModePerm)
	if err != nil {
		err = errors.New(fmt.Sprintf("Trace: WriteFile err %s filename %s", err, filename))
	}
	defer f.Close()
	n, err := f.Write(infos)
	if err != nil {
		err = errors.New(fmt.Sprintf("CPUProfile: WriteFile err %s filename %s num %d", err, filename, n))
	}
	return
}

func ProfileTraceToBytes(sec int) (ret []byte, err error) {
	buffer := bytes.NewBuffer([]byte{})
	if err = trace.Start(buffer); err != nil {
		// trace.Start failed, so no writes yet.
		err = errors.New(fmt.Sprintf("Trace: Could not enable tracing: %s", err))
		return
	}
	sleep(time.Duration(sec) * time.Second)
	trace.Stop()
	return buffer.Bytes(), nil
}
